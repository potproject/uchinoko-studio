package controller

import (
	"encoding/json"
	"sync"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/potproject/uchinoko-studio/api"
	"github.com/potproject/uchinoko-studio/data"
	"github.com/potproject/uchinoko-studio/db"
	"github.com/potproject/uchinoko-studio/envgen"
	"github.com/sashabaranov/go-openai"
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
	ChatResponseChunkOutputType = "chat-response-chunk"
	ErrorOutputType             = "error"
	FinishOutputType            = "finish"
)

func messageProcess(mt int, msg []byte, fileType string, apiKey string) (string, error) {
	if mt == websocket.BinaryMessage {
		return api.Whisper(apiKey, msg, fileType)
	}
	if mt == websocket.TextMessage {
		textInput := TextInput{}
		err := json.Unmarshal(msg, &textInput)
		if err != nil {
			return "", err
		}
		return textInput.Text, nil
	}
	return "", nil
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

func getTranscriptionApiKey() string {
	return envgen.Get().OPENAI_API_KEY()
}

func getChatApiKey(chatType string) string {
	if chatType == "openai" {
		return envgen.Get().OPENAI_API_KEY()
	}
	if chatType == "anthropic" {
		return envgen.Get().ANTHROPIC_API_KEY()
	}
	return ""
}

func getVoiceEndpoint(voiceType string) string {
	if voiceType == "voicevox" {
		return envgen.Get().VOICEVOX_ENDPOINT()
	}
	if voiceType == "bertvits2" {
		return envgen.Get().BERTVITS2_ENDPOINT()
	}
	if voiceType == "stylebertvits2" {
		return envgen.Get().STYLEBERTVIT2_ENDPOINT()
	}
	return ""
}

func WSTalk() fiber.Handler {
	return websocket.New(func(c *websocket.Conn) {
		id := c.Params("id")
		fileType := c.Params("fileType")
		characterId := c.Params("characterId")

		if fileType != "mp4" && fileType != "mp3" && fileType != "wav" && fileType != "webm" {
			c.Close()
			return
		}

		character, err := db.GetCharacterConfig(characterId)
		if err != nil {
			sendError(c, err)
			return
		}

		chatType := character.Chat.Type
		chatModel := character.Chat.Model
		chatSystemPropmt := character.Chat.SystemPrompt

		voice := character.Voice[0]

		format := "wav"

		wsSendTextMessage(c, ConnectionOutputType, format)

		for {
			mt, msg, err := c.ReadMessage()
			//start := time.Now()
			if err != nil {
				sendError(c, err)
				break
			}

			requestText, err := messageProcess(mt, msg, fileType, getTranscriptionApiKey())
			if err != nil {
				sendError(c, err)
				break
			}

			wsSendTextMessage(c, ChatRequestOutputType, requestText)

			var wg sync.WaitGroup
			chunkMessage := make(chan api.TextMessage)
			chunkAudio := make(chan api.AudioMessage)

			chatDone := make(chan bool)
			ttsDone := make(chan bool)

			// Chat処理
			wg.Add(1)
			go func() {
				err := runChatStream(id, requestText, chatType, getChatApiKey(chatType), chatSystemPropmt, chatModel, chunkMessage)
				if err != nil {
					sendError(c, err)
				}
				chatDone <- true
				wg.Done()
			}()

			// TTS処理
			wg.Add(1)
			go func() {
				err := runTTSStream(voice.Type, getVoiceEndpoint(voice.Type), voice.SpeakerID, voice.ModelID, voice.ModelFile, chunkMessage, chunkAudio, chatDone)
				if err != nil {
					sendError(c, err)
				}
				ttsDone <- true
				wg.Done()
			}()

			// WebSocketへの出力
			wg.Add(1)
			go func() {
				runWSSend(c, chunkAudio, ttsDone)
				wg.Done()
			}()
			wg.Wait()

			wsSendTextMessage(c, FinishOutputType, "")

			close(chunkMessage)
			close(chunkAudio)
			close(chatDone)
			close(ttsDone)
		}
	})
}

func runWSSend(c *websocket.Conn, outAudioMessage chan api.AudioMessage, ttsDone chan bool) {
	text := ""
	for {
		select {
		case a := <-outAudioMessage:
			if len(a.Audio) == 0 {
				continue
			}
			wsSendTextMessage(c, ChatResponseChunkOutputType, a.Text)
			text += a.Text
			wsSendBinaryMessage(c, a.Audio)
		case <-ttsDone:
			wsSendTextMessage(c, ChatResponseOutputType, text)
			return
		}
	}
}

func runTTSStream(voiceType string, endpoint string, voiceSpeaker string, voiceModel string, voiceModelFile string, chunkMessage chan api.TextMessage, outAudioMessage chan api.AudioMessage, chatDone chan bool) error {
	var err error
	if voiceType == "voicevox" {
		err = api.VoicevoxTTSStream(endpoint, voiceSpeaker, chunkMessage, outAudioMessage, chatDone)
	}
	if voiceType == "stylebertvits2" {
		err = api.StyleBertVits2TTSStream(endpoint, voiceModel, voiceModelFile, voiceSpeaker, chunkMessage, outAudioMessage, chatDone)
	}
	if voiceType == "bertvits2" {
		err = api.BertVits2TTSStream(endpoint, voiceModel, voiceSpeaker, chunkMessage, outAudioMessage, chatDone)
	}
	if err != nil {
		return err
	}
	return nil
}

func runChatStream(id string, requestText string, chatType string, apiKey string, chatSystemPropmt string, chatModel string, chunkMessage chan api.TextMessage) error {
	cm, _, err := db.GetChatMessage(id)
	if err != nil {
		return err
	}
	var ncm []openai.ChatCompletionMessage

	if chatType == "openai" {
		ncm, err = api.OpenAIChatStream(apiKey, chatSystemPropmt, chatModel, cm.Chat, requestText, chunkMessage)
	}
	if chatType == "anthropic" {
		ncm, err = api.AnthropicChatStream(apiKey, chatSystemPropmt, chatModel, cm.Chat, requestText, chunkMessage)
	}

	if err != nil {
		return err
	}

	err = db.PutChatMessage(id, data.ChatMessage{
		Chat: ncm,
	})

	if err != nil {
		return err
	}
	return nil
}

func sendError(c *websocket.Conn, err error) error {
	return wsSendTextMessage(c, ErrorOutputType, err.Error())
}
