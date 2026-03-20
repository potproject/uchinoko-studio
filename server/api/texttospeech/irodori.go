package texttospeech

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/potproject/uchinoko-studio/envgen"
)

const (
	irodoriDefaultCheckpoint = "Aratako/Irodori-TTS-500M"
	irodoriAPISuffix         = "/_run_generation"
)

var irodoriCallPrefixes = []string{"/call", "/gradio_api/call"}

type irodoriCallCreatedResponse struct {
	EventID string `json:"event_id"`
}

type irodoriCallOutput struct {
	Data []any
}

type irodoriFileData struct {
	Path     string `json:"path"`
	URL      string `json:"url"`
	OrigName string `json:"orig_name"`
	MimeType string `json:"mime_type"`
}

func irodori(endpoint string, checkpoint string, referenceAudioPath string, text string) ([]byte, error) {
	baseURL := strings.TrimRight(strings.TrimSpace(endpoint), "/")
	if baseURL == "" {
		return nil, fmt.Errorf("IRODORI_TTS_ENDPOINT is not configured")
	}

	if strings.TrimSpace(checkpoint) == "" {
		checkpoint = irodoriDefaultCheckpoint
	}

	refInput, err := buildIrodoriReferenceAudio(referenceAudioPath)
	if err != nil {
		return nil, err
	}

	payload := []any{
		checkpoint,
		"cuda",
		"bf16",
		"cuda",
		"bf16",
		false,
		text,
		refInput,
		40,
		1,
		"",
		"independent",
		3,
		5,
		"",
		0.5,
		1.0,
		true,
		"",
		"",
		"",
		false,
		"",
		"0.9",
		"",
	}

	client := &http.Client{Timeout: 3 * time.Minute}
	var lastErr error

	for _, prefix := range irodoriCallPrefixes {
		audio, callErr := irodoriCall(client, baseURL, prefix+irodoriAPISuffix, payload)
		if callErr == nil {
			return audio, nil
		}
		lastErr = callErr
	}

	return nil, lastErr
}

func buildIrodoriReferenceAudio(referenceAudioPath string) (any, error) {
	ref := strings.TrimSpace(referenceAudioPath)
	if ref == "" {
		return nil, nil
	}

	if isHTTPURL(ref) {
		return map[string]any{"path": ref}, nil
	}

	publicRefURL, err := buildLocalReferenceAudioURL(ref)
	if err != nil {
		return nil, err
	}

	return map[string]any{"path": publicRefURL}, nil
}

func buildLocalReferenceAudioURL(referenceAudioPath string) (string, error) {
	host := envgen.Get().HOST()
	switch host {
	case "", "0.0.0.0", "::":
		host = "127.0.0.1"
	}

	refPath := filepath.ToSlash(strings.TrimSpace(referenceAudioPath))
	refPath = strings.TrimPrefix(refPath, "/")
	refPath = strings.TrimPrefix(refPath, "refs/")
	refPath = path.Clean(refPath)
	if refPath == "." || refPath == "" {
		return "", fmt.Errorf("referenceAudioPath is invalid")
	}
	if strings.HasPrefix(refPath, "../") {
		return "", fmt.Errorf("referenceAudioPath must stay inside refs/")
	}

	return (&url.URL{
		Scheme: "http",
		Host:   host + ":" + strconv.FormatInt(int64(envgen.Get().PORT()), 10),
		Path:   "/refs/" + refPath,
	}).String(), nil
}

func irodoriCall(client *http.Client, baseURL string, callPath string, payload []any) ([]byte, error) {
	requestBody, err := json.Marshal(map[string]any{"data": payload})
	if err != nil {
		return nil, err
	}

	callURL := baseURL + callPath
	req, err := http.NewRequest("POST", callURL, bytes.NewReader(requestBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("irodori API not found: %s", callURL)
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("irodori call failed: %s %s", res.Status, strings.TrimSpace(string(body)))
	}

	var created irodoriCallCreatedResponse
	if err := json.NewDecoder(res.Body).Decode(&created); err != nil {
		return nil, err
	}
	if created.EventID == "" {
		return nil, fmt.Errorf("irodori event_id is empty")
	}

	resultURL := callURL + "/" + created.EventID
	resultReq, err := http.NewRequest("GET", resultURL, nil)
	if err != nil {
		return nil, err
	}
	resultRes, err := client.Do(resultReq)
	if err != nil {
		return nil, err
	}
	defer resultRes.Body.Close()

	if resultRes.StatusCode < 200 || resultRes.StatusCode >= 300 {
		body, _ := io.ReadAll(resultRes.Body)
		return nil, fmt.Errorf("irodori result failed: %s %s", resultRes.Status, strings.TrimSpace(string(body)))
	}

	output, err := parseIrodoriResult(resultRes.Body)
	if err != nil {
		return nil, err
	}

	audioURL, err := findIrodoriAudioURL(baseURL, output.Data)
	if err != nil {
		return nil, err
	}

	audioRes, err := client.Get(audioURL)
	if err != nil {
		return nil, err
	}
	defer audioRes.Body.Close()

	if audioRes.StatusCode < 200 || audioRes.StatusCode >= 300 {
		body, _ := io.ReadAll(audioRes.Body)
		return nil, fmt.Errorf("failed to fetch generated audio: %s %s", audioRes.Status, strings.TrimSpace(string(body)))
	}

	return io.ReadAll(audioRes.Body)
}

func parseIrodoriResult(r io.Reader) (*irodoriCallOutput, error) {
	body, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	var direct irodoriCallOutput
	if err := json.Unmarshal(body, &direct.Data); err == nil && len(direct.Data) > 0 {
		return &direct, nil
	}

	return parseIrodoriSSE(body)
}

func parseIrodoriSSE(body []byte) (*irodoriCallOutput, error) {
	scanner := bufio.NewScanner(bytes.NewReader(body))
	scanner.Buffer(make([]byte, 1024), 1024*1024)

	currentEvent := ""
	dataLines := []string{}

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			output, done, err := consumeIrodoriEvent(currentEvent, dataLines)
			if err != nil {
				return nil, err
			}
			if done {
				return output, nil
			}
			currentEvent = ""
			dataLines = nil
			continue
		}

		if strings.HasPrefix(line, "event:") {
			currentEvent = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
			continue
		}
		if strings.HasPrefix(line, "data:") {
			dataLines = append(dataLines, strings.TrimSpace(strings.TrimPrefix(line, "data:")))
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	output, done, err := consumeIrodoriEvent(currentEvent, dataLines)
	if err != nil {
		return nil, err
	}
	if done {
		return output, nil
	}

	return nil, fmt.Errorf("irodori generation did not return a completed result")
}

func consumeIrodoriEvent(event string, dataLines []string) (*irodoriCallOutput, bool, error) {
	if event == "" || len(dataLines) == 0 {
		return nil, false, nil
	}

	dataJSON := strings.Join(dataLines, "\n")

	if event == "error" {
		var message string
		_ = json.Unmarshal([]byte(dataJSON), &message)
		if message == "" {
			message = dataJSON
		}
		return nil, false, fmt.Errorf("irodori generation failed: %s", strings.TrimSpace(message))
	}
	if event != "complete" {
		return nil, false, nil
	}

	var output irodoriCallOutput
	if err := json.Unmarshal([]byte(dataJSON), &output.Data); err != nil {
		return nil, false, err
	}

	return &output, true, nil
}

func findIrodoriAudioURL(baseURL string, value any) (string, error) {
	switch v := value.(type) {
	case []any:
		for _, item := range v {
			audioURL, err := findIrodoriAudioURL(baseURL, item)
			if err == nil {
				return audioURL, nil
			}
		}
	case map[string]any:
		file := irodoriFileData{}
		if pathValue, ok := v["path"].(string); ok {
			file.Path = pathValue
		}
		if urlValue, ok := v["url"].(string); ok {
			file.URL = urlValue
		}
		if mimeType, ok := v["mime_type"].(string); ok {
			file.MimeType = mimeType
		}
		if origName, ok := v["orig_name"].(string); ok {
			file.OrigName = origName
		}

		if file.URL != "" || file.Path != "" {
			if file.MimeType == "" || strings.HasPrefix(file.MimeType, "audio/") || strings.HasSuffix(strings.ToLower(file.OrigName), ".wav") || strings.HasSuffix(strings.ToLower(file.Path), ".wav") {
				return resolveIrodoriFileURL(baseURL, file)
			}
		}

		for _, nested := range v {
			audioURL, err := findIrodoriAudioURL(baseURL, nested)
			if err == nil {
				return audioURL, nil
			}
		}
	}

	return "", fmt.Errorf("generated audio was not found in irodori response")
}

func resolveIrodoriFileURL(baseURL string, file irodoriFileData) (string, error) {
	if file.URL != "" {
		if isHTTPURL(file.URL) {
			return file.URL, nil
		}

		base, err := url.Parse(baseURL)
		if err != nil {
			return "", err
		}
		ref, err := url.Parse(file.URL)
		if err != nil {
			return "", err
		}
		return base.ResolveReference(ref).String(), nil
	}

	return "", fmt.Errorf("generated audio URL is missing")
}

func isHTTPURL(raw string) bool {
	u, err := url.Parse(raw)
	return err == nil && (u.Scheme == "http" || u.Scheme == "https")
}
