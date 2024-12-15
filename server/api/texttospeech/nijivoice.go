package texttospeech

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const nijiVoiceEndpoint = "https://api.nijivoice.com/api/platform/v1/"
const nijivoiceGenerateVoice = "voice-actors/{id}/generate-voice"

type NijivoiceRequest struct {
	Script string `json:"script"`
	Speed  string `json:"speed"`
	Format string `json:"format"`
}

type NojivoiceResponse struct {
	GeneratedVoice struct {
		AudioFileUrl         string `json:"audioFileUrl"`
		AudioFileDownloadUrl string `json:"audioFileDownloadUrl"`
		Duration             int    `json:"duration"`
		RemainingCredits     int    `json:"remainingCredits"`
	} `json:"generatedVoice"`
}

func nijivoice(apiKey string, id string, text string) ([]byte, error) {
	url := nijiVoiceEndpoint + strings.Replace(nijivoiceGenerateVoice, "{id}", id, 1)

	reqBody := NijivoiceRequest{
		Script: text,
		Speed:  "1.0",
		Format: "wav",
	}

	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqJSON))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", apiKey)
	client := new(http.Client)
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Failed to generate voice: %s", res.Status)
	}

	var response NojivoiceResponse
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, err
	}

	if response.GeneratedVoice.AudioFileDownloadUrl == "" {
		return nil, fmt.Errorf("No audio file download url")
	}

	audioRes, err := http.Get(response.GeneratedVoice.AudioFileDownloadUrl)
	if err != nil {
		return nil, err
	}
	defer audioRes.Body.Close()

	return io.ReadAll(audioRes.Body)
}
