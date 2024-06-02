package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
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

func messageProcess(mt int, msg []byte, language string, typeTranscription string, apiKey string) (text string, image *data.Image, err error) {
	if mt == websocket.BinaryMessage {
		discreteType, extension, err := detectBinaryFileType(msg)
		if err != nil {
			return "", nil, err
		}
		if discreteType == "audio" {
			if typeTranscription == "google_speech_to_text" {
				text, err = speechtotext.GoSpeech(apiKey, msg, extension, language)
				return text, nil, err
			}
			if typeTranscription == "openai_speech_to_text" {
				text, err = speechtotext.OpenAISpeech(apiKey, msg, extension, language)
				return text, nil, err
			}
			if typeTranscription == "vosk_server" {
				text, err = speechtotext.VoskServer(apiKey, msg, extension, language)
				return text, nil, err
			}
		}

		if discreteType == "image" {
			return "", &data.Image{
				Extension: extension,
				Data:      msg,
			}, nil
		}
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

// Support File Types: mp3, wav, webm, ogg, jpg, png
func WSTalk() fiber.Handler {
	return websocket.New(func(c *websocket.Conn) {
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

		voices := character.Voice

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
				tokens, err = runChatStream(id, voices, character.MultiVoice, requestText, requestImage, chatType, getChatApiKey(chatType), chatSystemPropmt, character.Chat.MaxHistory, chatModel, chunkMessage)
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
	}, websocket.Config{
		EnableCompression: true,
	})
}

func detectBinaryFileType(data []byte) (string, string, error) {
	if len(data) < 12 {
		return "", "", fmt.Errorf("file size is too small")
	}

	switch {
	case bytes.HasPrefix(data, []byte{0x4F, 0x67, 0x67, 0x53}):
		return "audio", "ogg", nil
	case bytes.HasPrefix(data, []byte{0x52, 0x49, 0x46, 0x46}) && bytes.Contains(data[:12], []byte{0x57, 0x41, 0x56, 0x45}):
		return "audio", "wav", nil
	case bytes.HasPrefix(data, []byte{0xFF, 0xFB, 0x90}) || bytes.HasPrefix(data, []byte{0xFF, 0xFA, 0x90}) || bytes.HasPrefix(data, []byte{0x49, 0x44, 0x33}):
		return "audio", "mp3", nil
	case bytes.HasPrefix(data, []byte{0x1A, 0x45, 0xDF, 0xA3}): // or mkv
		return "audio", "webm", nil
	case bytes.HasPrefix(data, []byte{0xFF, 0xD8}):
		return "image", "jpg", nil
	case bytes.HasPrefix(data, []byte{0x89, 0x50, 0x4E, 0x47}):
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
			return
		}
	}
}

func runChatStream(id string, voices []data.CharacterConfigVoice, multi bool, requestText string, requestImage *data.Image, chatType string, apiKey string, chatSystemPropmt string, maxHistory int64, chatModel string, chunkMessage chan api.ChunkMessage) (*data.Tokens, error) {
	var t *data.Tokens
	cm, _, err := db.GetChatMessage(id)
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

	ncm, t, err := chatStream(apiKey, voices, multi, chatSystemPropmt, chatModel, cm.Chat, requestText, requestImage, chunkMessage)
	if err != nil {
		return t, err
	}

	err = db.PutChatMessage(id, data.ChatMessage{
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
