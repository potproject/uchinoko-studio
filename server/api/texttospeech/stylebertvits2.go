package texttospeech

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

const styleBertVits2G2pEndpoint = "api/g2p"
const styleBertVits2SynthesisEndpoint = "api/synthesis"

type styleBertVits2G2pRequestBody struct {
	Text string `json:"text"`
}

type styleBertVits2SynthesisRequestBody struct {
	Model           string          `json:"model"`
	Speaker         string          `json:"speaker"`
	ModelFile       string          `json:"modelFile"`
	Text            string          `json:"text"`
	MoraToneList    json.RawMessage `json:"moraToneList"`
	AccentModified  bool            `json:"accentModified,omitempty"`
	Noise           float64         `json:"noise,omitempty"`
	NoiseW          float64         `json:"noisew,omitempty"`
	IntonationScale float64         `json:"intonationScale,omitempty"`
	PitchScale      float64         `json:"pitchScale,omitempty"`
	Speed           float64         `json:"speed,omitempty"`
	Style           string          `json:"style,omitempty"`
	StyleWeight     float64         `json:"styleWeight,omitempty"`
	SdpRatio        float64         `json:"sdpRatio,omitempty"`
	SilenceAfter    float64         `json:"silenceAfter,omitempty"`
}

func styleBertVits2TTS(endpoint string, model string, modelFile string, speaker string, text string) ([]byte, error) {
	client := new(http.Client)
	g2pQuery := endpoint + styleBertVits2G2pEndpoint
	g2pReqBody := styleBertVits2G2pRequestBody{Text: text}
	g2pReqBodyBytes, err := json.Marshal(g2pReqBody)
	if err != nil {
		return nil, err
	}
	g2pReq, err := http.NewRequest("POST", g2pQuery, bytes.NewBuffer(g2pReqBodyBytes))
	if err != nil {
		return nil, err
	}
	g2pReq.Header.Set("Content-Type", "application/json")
	g2pReq.Header.Set("Accept", "application/json")
	g2pRes, err := client.Do(g2pReq)
	if err != nil {
		return nil, err
	}
	defer g2pRes.Body.Close()

	g2pResBodyBytes, err := io.ReadAll(g2pRes.Body)
	if err != nil {
		return nil, err
	}

	synthesisQuery := endpoint + styleBertVits2SynthesisEndpoint
	synthesisReqBody := styleBertVits2SynthesisRequestBody{
		Text:            text,
		MoraToneList:    g2pResBodyBytes,
		Model:           model,
		ModelFile:       modelFile,
		Speaker:         speaker,
		Style:           "Neutral",
		StyleWeight:     5,
		Speed:           1.0,
		SdpRatio:        0.2,
		Noise:           0.6,
		NoiseW:          0.8,
		IntonationScale: 1.0,
		PitchScale:      1.0,
		SilenceAfter:    0.5,
		AccentModified:  false,
	}
	synthesisReqBodyBytes, err := json.Marshal(synthesisReqBody)
	if err != nil {
		return nil, err
	}
	synthesisReq, err := http.NewRequest("POST", synthesisQuery, bytes.NewBuffer(synthesisReqBodyBytes))
	if err != nil {
		return nil, err
	}
	synthesisReq.Header.Set("Content-Type", "application/json")
	synthesisReq.Header.Set("Accept", "audio/wav")
	synthesisRes, err := client.Do(synthesisReq)
	if err != nil {
		return nil, err
	}
	defer synthesisRes.Body.Close()

	return io.ReadAll(synthesisRes.Body)
}
