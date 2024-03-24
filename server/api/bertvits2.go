package api

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/url"
)

const bertvits2VoiceEndpoint = "voice"

func BertVits2TTSStream(endpoint string, modelId string, speakerId string, chunkMessage <-chan TextMessage, outAudio chan []byte, outText chan string) error {
	for {
		select {
		case t := <-chunkMessage:
			if len(t.Text) == 0 {
				if t.IsFinal {
					return nil
				}
				continue
			}
			bin, err := bertVits2TTS(endpoint, modelId, speakerId, t.Text)
			if err != nil {
				log.Printf("Error: %s", err.Error())
				return err
			}
			outText <- t.Text
			outAudio <- bin
			if t.IsFinal {
				return nil
			}
		}
	}
}

func bertVits2TTS(endpoint string, modelId string, speakerId string, text string) ([]byte, error) {
	client := new(http.Client)

	voiceQuery := endpoint + bertvits2VoiceEndpoint + "?model_id=" + modelId + "&speaker_id=" + speakerId + "&sdp_ratio=0.2&noise=0.1&noisew=1&length=0.9&language=JP&auto_translate=false&auto_split=false&emotion=0"
	textForm := url.Values{}
	textForm.Add("text", text)
	vReq, err := http.NewRequest("POST", voiceQuery, bytes.NewBufferString(textForm.Encode()))
	vReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	vReq.Header.Set("Accept", "audio/wav")

	if err != nil {
		return nil, err
	}

	vRes, err := client.Do(vReq)
	if err != nil {
		return nil, err
	}
	defer vRes.Body.Close()

	return io.ReadAll(vRes.Body)

}
