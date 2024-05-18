package controller

import (
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

func messageProcess(mt int, msg []byte, fileType string, language string, typeTranscription string, apiKey string) (string, error) {
	if mt == websocket.BinaryMessage {
		if typeTranscription == "google_speech_to_text" {
			return speechtotext.GoSpeech(apiKey, msg, fileType, language)
		}
		if typeTranscription == "openai_speech_to_text" {
			return speechtotext.OpenAISpeech(apiKey, msg, fileType, language)
		}
		return "", fmt.Errorf("unsupported transcription type: %s", typeTranscription)
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

func getTranscriptionApiKey(transcriptionType string) string {
	if transcriptionType == "google_speech_to_text" {
		return envgen.Get().GOOGLE_SPEECH_TO_TEXT_API_KEY()
	}
	if transcriptionType == "openai_speech_to_text" {
		return envgen.Get().OPENAI_SPEECH_TO_TEXT_API_KEY()
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

// Support File Types: mp3, wav, webm, ogg
func WSTalk() fiber.Handler {
	return websocket.New(func(c *websocket.Conn) {
		id := c.Params("id")
		fileType := c.Params("fileType")
		characterId := c.Params("characterId")

		if fileType != "mp3" && fileType != "wav" && fileType != "webm" && fileType != "ogg" {
			c.Close()
			return
		}

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

			requestText, err := messageProcess(mt, msg, fileType, general.Language, general.Transcription.Type, getTranscriptionApiKey(general.Transcription.Type))
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

			// Chat処理
			wg.Add(1)
			go func() {
				err := runChatStream(id, voices, character.MultiVoice, requestText, chatType, getChatApiKey(chatType), chatSystemPropmt, chatModel, chunkMessage)
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

			wsSendTextMessage(c, FinishOutputType, "")

			close(chunkMessage)
			close(chunkAudio)
			close(chatDone)
			close(ttsDone)
		}
	})
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

func runChatStream(id string, voices []data.CharacterConfigVoice, multi bool, requestText string, chatType string, apiKey string, chatSystemPropmt string, chatModel string, chunkMessage chan api.ChunkMessage) error {
	cm, _, err := db.GetChatMessage(id)
	if err != nil {
		return err
	}

	var chatStream chat.ChatStream
	if chatType == "openai" {
		chatStream = chat.OpenAIChatStream
	}
	if chatType == "anthropic" {
		chatStream = chat.AnthropicChatStream
	}
	if chatType == "cohere" {
		chatStream = chat.CohereChatStream
	}
	if chatType == "gemini" {
		chatStream = chat.GeminiChatStream
	}
	if chatType == "openai-local" {
		chatStream = chat.OpenAILocalChatStream
	}

	ncm, err := chatStream(apiKey, voices, multi, chatSystemPropmt, chatModel, cm.Chat, requestText, chunkMessage)
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
