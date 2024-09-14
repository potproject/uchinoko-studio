package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"sync"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/potproject/uchinoko-studio/api"
	"github.com/potproject/uchinoko-studio/api/chat"
	"github.com/potproject/uchinoko-studio/api/speechtotext"
	"github.com/potproject/uchinoko-studio/api/texttospeech"
	"github.com/potproject/uchinoko-studio/data"
	"github.com/potproject/uchinoko-studio/db"
	"github.com/potproject/uchinoko-studio/envgen"
	"github.com/potproject/uchinoko-studio/osc"
)

type TextInput struct {
	Text string `json:"text"`
}

type TextOutput struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

const (
	ConnectionOutputType        = "connection"
	ChatRequestOutputType       = "chat-request"
	ChatResponseOutputType      = "chat-response"
	ChatResponseChangeCharacter = "chat-response-change-character"
	ChatResponseChangeBehavior  = "chat-response-change-behavior"
	ChatResponseChunkOutputType = "chat-response-chunk"
	ErrorOutputType             = "error"
	FinishOutputType            = "finish"
)

type MultipartMessage struct {
	Parts []MultipartMessagePart `json:"parts"`
}

type MultipartMessagePart struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content-type"`
	Data        []byte `json:"data"`
}

func parseMultipartMessage(message []byte) ([]MultipartMessagePart, error) {
	boundary := "boundaryUchinoko"
	reader := multipart.NewReader(bytes.NewReader(message), boundary)
	var parts []MultipartMessagePart

	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		buf := new(bytes.Buffer)
		buf.ReadFrom(part)
		fileData := MultipartMessagePart{
			Filename:    part.FileName(),
			ContentType: part.Header.Get("Content-Type"),
			Data:        buf.Bytes(),
		}

		parts = append(parts, fileData)
	}

	return parts, nil
}

func messageProcess(mt int, msg []byte, language string, typeTranscription string, apiKey string) (text string, image *data.Image, err error) {
	if mt == websocket.BinaryMessage {
		mm, err := parseMultipartMessage(msg)
		if err != nil {
			return "", nil, err
		}

		var text string
		var image *data.Image
		for _, m := range mm {
			discreteType, extension, err := detectBinaryFileType(m.ContentType)
			if err != nil {
				return "", nil, err
			}

			switch true {
			case discreteType == "image":
				image = &data.Image{
					Extension: extension,
					Data:      m.Data,
				}
			case discreteType == "audio" && typeTranscription == "google_speech_to_text":
				text, err = speechtotext.GoSpeech(apiKey, m.Data, extension, language)
			case discreteType == "audio" && typeTranscription == "openai_speech_to_text":
				text, err = speechtotext.OpenAISpeech(apiKey, m.Data, extension, language)
			case discreteType == "audio" && typeTranscription == "vosk_server":
				text, err = speechtotext.VoskServer(apiKey, m.Data, extension, language)
			}
			if err != nil {
				return "", nil, err
			}
		}
		return text, image, err
	}
	if mt == websocket.TextMessage {
		textInput := TextInput{}
		err := json.Unmarshal(msg, &textInput)
		if err != nil {
			return "", nil, err
		}
		return textInput.Text, nil, err
	}
	return
}

func wsSendTextMessage(c *websocket.Conn, msgType string, text string) error {
	chatResOutput, _ := json.Marshal(TextOutput{
		Type: msgType,
		Text: text,
	})
	return c.WriteMessage(websocket.TextMessage, chatResOutput)
}

func wsSendBinaryMessage(c *websocket.Conn, data []byte) error {
	return c.WriteMessage(websocket.BinaryMessage, data)
}

func getTranscriptionApiKey(transcriptionType string) string {
	if transcriptionType == "google_speech_to_text" {
		return envgen.Get().GOOGLE_SPEECH_TO_TEXT_API_KEY()
	}
	if transcriptionType == "openai_speech_to_text" {
		return envgen.Get().OPENAI_SPEECH_TO_TEXT_API_KEY()
	}
	if transcriptionType == "vosk_server" {
		return envgen.Get().VOSK_SERVER_ENDPOINT()
	}
	return ""
}

func getChatApiKey(chatType string) string {
	if chatType == "openai" {
		return envgen.Get().OPENAI_API_KEY()
	}
	if chatType == "anthropic" {
		return envgen.Get().ANTHROPIC_API_KEY()
	}
	if chatType == "cohere" {
		return envgen.Get().COHERE_API_KEY()
	}
	if chatType == "gemini" {
		return envgen.Get().GEMINI_API_KEY()
	}
	if chatType == "openai-local" {
		return envgen.Get().OPENAI_LOCAL_API_KEY()
	}
	return ""
}

func WSTalkCompressed() fiber.Handler {
	return websocket.New(WSTalk(), websocket.Config{
		EnableCompression: true,
	})
}

func WSTalkPlain() fiber.Handler {
	return websocket.New(WSTalk())
}

// Support File Types: mp3, wav, webm, ogg, jpg, png
func WSTalk() func(*websocket.Conn) {
	return func(c *websocket.Conn) {
		id := c.Params("id")
		characterId := c.Params("characterId")

		character, err := db.GetCharacterConfig(characterId)
		if err != nil {
			sendError(c, err)
			return
		}

		general, err := db.GetGeneralConfig()
		if err != nil {
			sendError(c, err)
			return
		}

		chatType := character.Chat.Type
		chatModel := character.Chat.Model
		chatSystemPropmt := character.Chat.SystemPrompt

		format := "wav"

		wsSendTextMessage(c, ConnectionOutputType, format)

		for {
			mt, msg, err := c.ReadMessage()
			if err != nil {
				sendError(c, err)
				break
			}

			allow, err := db.RateLimitIsAllowed(id, character.Chat.Limit)
			if err != nil {
				sendError(c, err)
				return
			}
			if !allow {
				sendError(c, fmt.Errorf("rate limit exceeded"))
				return
			}

			requestText, requestImage, err := messageProcess(mt, msg, general.Language, general.Transcription.Type, getTranscriptionApiKey(general.Transcription.Type))
			if err != nil {
				sendError(c, err)
				break
			}

			wsSendTextMessage(c, ChatRequestOutputType, requestText)

			var wg sync.WaitGroup
			chunkMessage := make(chan api.ChunkMessage)
			chunkAudio := make(chan api.AudioMessage)
			changeVoice := make(chan data.CharacterConfigVoice)
			changeBehavior := make(chan data.CharacterConfigVoiceBehavior)

			chatDone := make(chan bool)
			ttsDone := make(chan bool)
			var tokens *data.Tokens

			// Chat処理
			wg.Add(1)
			go func() {
				var err error
				tokens, err = runChatStream(id, character, character.MultiVoice, requestText, requestImage, chatType, getChatApiKey(chatType), chatSystemPropmt, character.Chat.MaxHistory, chatModel, chunkMessage)
				if err != nil {
					sendError(c, err)
				}
				var totalToken int64
				if tokens != nil {
					totalToken = (tokens.InputTokens) + (tokens.OutputTokens)
				}
				err = db.AddRateLimit(id, 1, int64(totalToken))
				if err != nil {
					sendError(c, err)
				}
				chatDone <- true
				wg.Done()
			}()

			// TTS処理
			wg.Add(1)
			go func() {
				err := texttospeech.TTSStream(general, chunkMessage, changeVoice, changeBehavior, chunkAudio, chatDone)
				if err != nil {
					sendError(c, err)
				}
				ttsDone <- true
				wg.Done()
			}()

			// WebSocketへの出力
			wg.Add(1)
			go func() {
				runWSSend(c, chunkAudio, changeVoice, changeBehavior, ttsDone)
				wg.Done()
			}()
			wg.Wait()

			jsonTokens, _ := json.Marshal(tokens)
			wsSendTextMessage(c, FinishOutputType, string(jsonTokens))

			close(chunkMessage)
			close(chunkAudio)
			close(chatDone)
			close(ttsDone)
		}
	}
}

func detectBinaryFileType(contentType string) (string, string, error) {
	switch contentType {
	case "audio/mpeg":
		return "audio", "mp3", nil
	case "audio/wav":
		return "audio", "wav", nil
	case "audio/webm":
		return "audio", "webm", nil
	case "audio/ogg":
		return "audio", "ogg", nil
	case "image/jpeg":
		return "image", "jpg", nil
	case "image/png":
		return "image", "png", nil
	default:
		return "", "", fmt.Errorf("unsupported file type")
	}
}

func runWSSend(c *websocket.Conn, outAudioMessage chan api.AudioMessage, changeVoice chan data.CharacterConfigVoice, changeBehavior chan data.CharacterConfigVoiceBehavior, ttsDone chan bool) {
	text := ""
	for {
		select {
		case a := <-outAudioMessage:
			if len(a.Text) == 0 {
				continue
			}
			wsSendTextMessage(c, ChatResponseChunkOutputType, a.Text)
			text += a.Text
			if a.Audio != nil {
				wsSendBinaryMessage(c, *a.Audio)
			}
		case v := <-changeVoice:
			wsSendTextMessage(c, ChatResponseChangeCharacter, v.Identification)
		case b := <-changeBehavior:
			wsSendTextMessage(c, ChatResponseChangeBehavior, b.ImagePath)
		case <-ttsDone:
			wsSendTextMessage(c, ChatResponseOutputType, text)
			if envgen.Get().VRCHAT_OSC_ENABLED() {
				err := osc.SendChatBoxMessage(envgen.Get().VRCHAT_OSC_HOST(), envgen.Get().VRCHAT_OSC_PORT(), text)
				if err != nil {
					sendError(c, err)
				}
			}
			return
		}
	}
}

func runChatStream(id string, characterConfig data.CharacterConfig, multi bool, requestText string, requestImage *data.Image, chatType string, apiKey string, chatSystemPropmt string, maxHistory int64, chatModel string, chunkMessage chan api.ChunkMessage) (*data.Tokens, error) {
	var t *data.Tokens
	cm, _, err := db.GetChatMessage(id, characterConfig.General.ID)
	if err != nil {
		return t, err
	}

	if maxHistory > 0 && int64(len(cm.Chat)) > maxHistory {
		cm.Chat = cm.Chat[len(cm.Chat)-int(maxHistory):]
	}

	var chatStream chat.ChatStream
	// image support: ok
	if chatType == "openai" {
		chatStream = chat.OpenAIChatStream
	}
	// image support: ok
	if chatType == "anthropic" {
		chatStream = chat.AnthropicChatStream
	}
	// image support: ng
	if chatType == "cohere" {
		chatStream = chat.CohereChatStream
	}
	// image support: ok
	if chatType == "gemini" {
		chatStream = chat.GeminiChatStream
	}
	// image support: unknown
	if chatType == "openai-local" {
		chatStream = chat.OpenAILocalChatStream
	}

	ncm, t, err := chatStream(apiKey, characterConfig.Voice, multi, chatSystemPropmt, chatModel, cm.Chat, requestText, requestImage, chunkMessage)
	if err != nil {
		return t, err
	}

	err = db.PutChatMessage(id, characterConfig.General.ID,
		data.ChatMessage{
			Chat: ncm,
		})

	if err != nil {
		return t, err
	}
	return t, nil
}

func sendError(c *websocket.Conn, err error) error {
	return wsSendTextMessage(c, ErrorOutputType, err.Error())
}
