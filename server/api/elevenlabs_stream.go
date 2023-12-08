package api

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"strings"

	"github.com/gorilla/websocket"
)

const elevenlabsWebsocketURL = "wss://api.elevenlabs.io/v1/text-to-speech/{voice_id}/stream-input?model_id={model}&optimize_streaming_latency={optimize_streaming_latency}&output_format={output_format}"

type ElevenLabsWebsocketRequestStart struct {
	Text             string                                     `json:"text"`
	VoiceSettings    ElevenLabsWebsocketRequestVoiceSettings    `json:"voice_settings"`
	GenerationConfig ElevenLabsWebsocketRequestGenerationConfig `json:"generation_config"`
	XiApiKey         string                                     `json:"xi_api_key"`
}

type ElevenLabsWebsocketRequestVoiceSettings struct {
	Stability       float64 `json:"stability"`
	SimilarityBoost bool    `json:"similarity_boost"`
}

type ElevenLabsWebsocketRequestGenerationConfig struct {
	ChunkLengthSchedule []int `json:"chunk_length_schedule"`
}

type ElevenLabsWebsocketRequestChunk struct {
	Text string `json:"text"`
}

type ElevenLabsWebsocketRequestEnd struct {
	Text string `json:"text"` // Should always be an empty string "".
}

type ElevenLabsWebsocketResponse struct {
	Audio               string `json:"audio"`
	IsFinal             bool   `json:"isFinal"`
	NormalizedAlignment struct {
		CharStartTimesMS []int    `json:"char_start_times_ms"`
		CharsDurationsMS []int    `json:"chars_durations_ms"`
		Chars            []string `json:"chars"`
	} `json:"normalizedAlignment"`
}

func ElevenLabsTTSWebsocket(c *ElevenLabsClientExtend, voiceID string, outputFormat string, chunkMessage <-chan TextMessage, outAudio chan []byte, outText chan string, done chan bool) error {
	// Setup
	url := elevenlabsWebsocketURL
	url = strings.Replace(url, "{voice_id}", voiceID, 1)
	url = strings.Replace(url, "{model}", elevenlabsModelID, 1)
	url = strings.Replace(url, "{optimize_streaming_latency}", "3", 1)
	url = strings.Replace(url, "{output_format}", outputFormat, 1)

	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Println("dial error:", err)
		return err
	}

	// Read Routine
	go func() {
		for {
			res := ElevenLabsWebsocketResponse{}
			_, bytes, err := ws.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					log.Printf("read error: %v", err)
				}
				done <- true
				return
			}

			json.Unmarshal(bytes, &res)
			// base64 decode
			decudeAudio, err := base64.StdEncoding.DecodeString(res.Audio)
			if err != nil {
				log.Printf("base64 decode error: %v", err)
				done <- true
				return
			}
			outAudio <- decudeAudio
		}
	}()

	// Write Routine
	go func() {
		for {
			select {
			case t := <-chunkMessage:
				outText <- t.Text
				if t.IsFirst {
					err := ws.WriteJSON(ElevenLabsWebsocketRequestStart{
						Text: t.Text,
						VoiceSettings: ElevenLabsWebsocketRequestVoiceSettings{
							Stability:       0.8,
							SimilarityBoost: true,
						},
						GenerationConfig: ElevenLabsWebsocketRequestGenerationConfig{
							ChunkLengthSchedule: []int{120, 160, 250, 290},
						},
						XiApiKey: c.ApiKey,
					})

					if err != nil {
						log.Printf("write error: %v", err)
						done <- true
						return
					}
					continue
				}

				err = ws.WriteJSON(ElevenLabsWebsocketRequestChunk{
					Text: t.Text,
				})

				if err != nil {
					log.Printf("write error: %v", err)
					done <- true
					return
				}

				if t.IsFinal {
					err := ws.WriteJSON(ElevenLabsWebsocketRequestEnd{
						Text: "",
					})

					if err != nil {
						log.Printf("write error: %v", err)
						done <- true
						return
					}
				}
			}
		}
	}()
	<-done
	ws.Close()
	return nil
}
