package api

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

const OpenAIwhisperEndpoint = "https://api.openai.com/v1/audio/transcriptions"

func Whisper(apiKey string, fileData []byte, extention string) (string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("model", "whisper-1")
	writer.WriteField("language", "ja")
	writer.WriteField("response_format", "text")
	part, err := writer.CreateFormFile("file", "audio."+extention)
	if err != nil {
		return "", err
	}
	part.Write(fileData)
	writer.Close()

	req, err := http.NewRequest("POST", OpenAIwhisperEndpoint, body)
	if err != nil {
		return "", err
	}

	req.Header.Add("Authorization", "Bearer "+apiKey)
	req.Header.Add("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad status: %s", resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
