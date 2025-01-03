package texttospeech

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
)

const voicevoxAudioQueryEndpoint = "audio_query"
const voicevoxSynthesisEndpoint = "synthesis"

func voicevox(endpoint string, speaker string, text string) ([]byte, error) {
	client := new(http.Client)
	audioQuery := endpoint + voicevoxAudioQueryEndpoint + "?speaker=" + speaker + "&text=" + url.QueryEscape(text)

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

	synthesis := endpoint + voicevoxSynthesisEndpoint + "?speaker=" + speaker
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
