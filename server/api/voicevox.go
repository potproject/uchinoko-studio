package api

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
)

const apiAudioQueryEndpoint = "audio_query"
const apiSynthesisEndpoint = "synthesis"

type Request struct {
	Text string `json:"text"`
}

func VoicevoxTTSStream(endpoint string, speaker int64, chunkMessage <-chan TextMessage, outAudio chan []byte, outText chan string, done chan bool) error {
	for {
		select {
		case t := <-chunkMessage:
			if len(t.Text) == 0 {
				if t.IsFinal {
					done <- true
					return nil
				}
				continue
			}
			bin, err := voicevoxTTS(endpoint, strconv.FormatInt(speaker, 10), t.Text)
			if err != nil {
				log.Printf("Error: %s", err.Error())
				return err
			}
			outText <- t.Text
			outAudio <- bin
			if t.IsFinal {
				done <- true
				return nil
			}
		}
	}
}

func voicevoxTTS(endpoint string, speaker string, text string) ([]byte, error) {
	client := new(http.Client)
	audioQuery := endpoint + apiAudioQueryEndpoint + "?speaker=" + speaker + "&text=" + url.QueryEscape(text)

	queryReq, err := http.NewRequest("POST", audioQuery, nil)
	if err != nil {
		return nil, err
	}
	queryRes, err := client.Do(queryReq)
	if err != nil {
		return nil, err
	}
	defer queryRes.Body.Close()

	qbin, err := io.ReadAll(queryRes.Body)
	if err != nil {
		return nil, err
	}

	synthesis := endpoint + apiSynthesisEndpoint + "?speaker=" + speaker
	synthesisReq, err := http.NewRequest("POST", synthesis, bytes.NewReader(qbin))

	if err != nil {
		return nil, err
	}

	synthesisReq.Header.Set("Accept", "audio/wav")
	synthesisReq.Header.Set("Content-Type", "application/json")

	synthesisRes, err := client.Do(synthesisReq)
	if err != nil {
		return nil, err
	}
	defer synthesisRes.Body.Close()

	return io.ReadAll(synthesisRes.Body)
}
