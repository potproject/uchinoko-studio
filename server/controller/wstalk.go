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

type BinaryInput struct {
	Data []byte
}

type TextInput struct {
	Text string `json:"text"`
}

type ConnectionOutput struct {
	BaseOutput // Type: connection
}

type BaseOutput struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type ChatRequestOutput struct {
	BaseOutput // Type: chat-request
}

type ChatResponseOutput struct {
	BaseOutput // Type: chat-response
}

type ChatResponseChunkOutput struct {
	BaseOutput // Type: chat-response-chunk
}

type ErrorOutput struct {
	BaseOutput // Type: error
}

type FinishOutput struct {
	BaseOutput // Type: finish
}

type BinaryOutput struct {
	Data []byte
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

		if voiceType != "voicevox" && voiceType != "elevenlabs" && voiceType != "bertvits2" {
			c.Close()
			return
		}

		openai := api.OpenAINewClient()

		el := api.ElevenLabsNewClient()
		voiceID := envgen.Get().ELEVENLABS_VOICEID()
		outputFormat := envgen.Get().ELEVENLABS_OUTPUT_FORMAT()

		voicevox := envgen.Get().VOICEVOX_ENDPOINT()
		voicevoxSpeaker := envgen.Get().VOICEVOX_SPEAKER()

		bertvits2 := envgen.Get().BERTVITS2_ENDPOINT()
		bertvits2ModelID := envgen.Get().BERTVITS2_MODEL_ID()
		bertvits2SpeakerId := envgen.Get().BERTVITS2_SPEAKER_ID()

		format := "wav"
		if voiceType == "elevenlabs" {
			format = outputFormat
		}

		connectionOutput, _ := json.Marshal(ConnectionOutput{
			BaseOutput: BaseOutput{
				Type: "connection",
				Text: format,
			},
		})

		c.WriteMessage(websocket.TextMessage, []byte(connectionOutput))

		for {
			mt, msg, err := c.ReadMessage()
			start := time.Now()
			if err != nil {
				sendError(c, err)
				break
			}

			var requestText string
			if mt == websocket.BinaryMessage {
				binaryInput := BinaryInput{
					Data: msg,
				}
				requestText, err = api.Whisper(openai, binaryInput.Data, fileType)
				if err != nil {
					sendError(c, err)
					break
				}
			}
			if mt == websocket.TextMessage {
				textInput := TextInput{}
				err = json.Unmarshal(msg, &textInput)
				if err != nil {
					sendError(c, err)
					break
				}
				requestText = textInput.Text
			}

			chatReqOutput, _ := json.Marshal(ChatRequestOutput{
				BaseOutput: BaseOutput{
					Type: "chat-request",
					Text: requestText,
				},
			})
			c.WriteMessage(websocket.TextMessage, chatReqOutput)

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
				ncm, err := api.ChatStream(openai, cm.Chat, requestText, chunkMessage, chatDone)
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
			} else if voiceType == "elevenlabs" {
				go func() {
					err := api.ElevenLabsTTSWebsocket(el, voiceID, outputFormat, chunkMessage, outAudio, outText)
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
					chatResOutput, _ := json.Marshal(ChatResponseOutput{
						BaseOutput: BaseOutput{
							Type: "chat-response-chunk",
							Text: t,
						},
					})
					c.WriteMessage(websocket.TextMessage, chatResOutput)
				case a := <-outAudio:
					if len(a) == 0 {
						continue
					}
					// バイナリ送信
					if firstSend {
						firstSend = false
						log.Printf("First Send: %s", time.Since(start))
					}
					binaryOutput := BinaryOutput{
						Data: a,
					}
					c.WriteMessage(websocket.BinaryMessage, binaryOutput.Data)
				case <-ttsDone:
					finishOutput, _ := json.Marshal(FinishOutput{
						BaseOutput: BaseOutput{
							Type: "finish",
							Text: "",
						},
					})
					c.WriteMessage(websocket.TextMessage, finishOutput)
					break Process
				case responseText := <-chatDone:
					chatResOutput, _ := json.Marshal(ChatResponseOutput{
						BaseOutput: BaseOutput{
							Type: "chat-response",
							Text: responseText,
						},
					})
					c.WriteMessage(websocket.TextMessage, chatResOutput)
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

func sendError(c *websocket.Conn, err error) {
	errorOutput, _ := json.Marshal(ErrorOutput{
		BaseOutput: BaseOutput{
			Type: "error",
			Text: err.Error(),
		},
	})
	c.WriteMessage(websocket.TextMessage, []byte(errorOutput))
}
