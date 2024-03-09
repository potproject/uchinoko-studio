package controller

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/potproject/uchinoko-studio/api"
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

func WSTalk() fiber.Handler {
	return websocket.New(func(c *websocket.Conn) {
		id := c.Params("id")
		fileType := c.Params("fileType")
		voiceType := c.Params("voiceType")

		if fileType != "mp4" && fileType != "mp3" && fileType != "wav" && fileType != "webm" {
			c.Close()
			return
		}

		if voiceType != "voicevox" && voiceType != "bertvits2" {
			c.Close()
			return
		}

		transcriptionsApiKey := envgen.Get().TRANSCRIPTIONS_API_KEY()
		chatEndpoint := envgen.Get().CHAT_ENDPOINT()
		chatApiKey := envgen.Get().CHAT_API_KEY()

		voicevox := envgen.Get().VOICEVOX_ENDPOINT()
		voicevoxSpeaker := envgen.Get().VOICEVOX_SPEAKER()

		bertvits2 := envgen.Get().BERTVITS2_ENDPOINT()
		bertvits2ModelID := envgen.Get().BERTVITS2_MODEL_ID()
		bertvits2SpeakerId := envgen.Get().BERTVITS2_SPEAKER_ID()

		format := "wav"

		wsSendTextMessage(c, ConnectionOutputType, format)

		for {
			mt, msg, err := c.ReadMessage()
			start := time.Now()
			if err != nil {
				sendError(c, err)
				break
			}

			requestText, err := messageProcess(mt, msg, fileType, transcriptionsApiKey)

			wsSendTextMessage(c, ChatRequestOutputType, requestText)

			outAudio := make(chan []byte)
			outText := make(chan string)

			chatDone := make(chan string)
			ttsDone := make(chan bool)
			chunkMessage := make(chan api.TextMessage)

			go func() {
				cm, _, err := db.GetChatMessage(id)
				if err != nil {
					sendError(c, err)
				}
				ncm, err := api.ChatStream(chatEndpoint, chatApiKey, cm.Chat, requestText, chunkMessage, chatDone)
				if err != nil {
					sendError(c, err)
				}
				db.PutChatMessage(id, data.ChatMessage{
					Chat: ncm,
				})
				if err != nil {
					sendError(c, err)
				}
			}()
			if voiceType == "voicevox" {
				go func() {
					err = api.VoicevoxTTSStream(voicevox, voicevoxSpeaker, chunkMessage, outAudio, outText)
					ttsDone <- true
					if err != nil {
						sendError(c, err)
					}
				}()
			} else {
				go func() {
					err := api.BertVits2TTSStream(bertvits2, bertvits2ModelID, bertvits2SpeakerId, chunkMessage, outAudio, outText)
					ttsDone <- true
					if err != nil {
						sendError(c, err)
					}
				}()
			}

			firstSend := true
		Process:
			for {
				select {
				case t := <-outText:
					if len(t) == 0 {
						continue
					}
					wsSendTextMessage(c, ChatResponseChunkOutputType, t)
				case a := <-outAudio:
					if len(a) == 0 {
						continue
					}
					// バイナリ送信
					if firstSend {
						firstSend = false
						log.Printf("First Send: %s", time.Since(start))
					}
					wsSendBinaryMessage(c, a)
				case <-ttsDone:
					wsSendTextMessage(c, FinishOutputType, "")
					break Process
				case responseText := <-chatDone:
					wsSendTextMessage(c, ChatResponseOutputType, responseText)
				}
			}
			close(outText)
			close(outAudio)
			close(chatDone)
			close(ttsDone)
			close(chunkMessage)
		}
	})
}

func sendError(c *websocket.Conn, err error) error {
	return wsSendTextMessage(c, ErrorOutputType, err.Error())
}
