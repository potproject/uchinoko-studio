package texttospeech

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
)

const bertvits2VoiceEndpoint = "voice"

func bertVits2(endpoint string, modelId string, speakerId string, text string) ([]byte, error) {
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
