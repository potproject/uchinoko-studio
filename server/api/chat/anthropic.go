package chat

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/potproject/uchinoko-studio/api"
	"github.com/potproject/uchinoko-studio/data"
)

const (
	anthropicMessagesEndpoint = "https://api.anthropic.com/v1/messages"
	anthropicVersion          = "2023-06-01"
)

type anthropicMessagesRequest struct {
	Model       string               `json:"model"`
	Messages    []anthropicMessage   `json:"messages"`
	System      []anthropicTextBlock `json:"system,omitempty"`
	MaxTokens   int                  `json:"max_tokens"`
	Stream      bool                 `json:"stream"`
	Temperature *float64             `json:"temperature,omitempty"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content any    `json:"content"`
}

type anthropicTextBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type anthropicImageBlock struct {
	Type   string               `json:"type"`
	Source anthropicImageSource `json:"source"`
}

type anthropicImageSource struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type"`
	Data      string `json:"data"`
}

type anthropicErrorResponse struct {
	Type  string `json:"type"`
	Error struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

type anthropicStreamMessageStartEvent struct {
	Message struct {
		Usage anthropicUsage `json:"usage"`
	} `json:"message"`
}

type anthropicStreamContentBlockDeltaEvent struct {
	Delta struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"delta"`
}

type anthropicStreamMessageDeltaEvent struct {
	Usage anthropicUsage `json:"usage"`
}

type anthropicUsage struct {
	InputTokens  int64 `json:"input_tokens"`
	OutputTokens int64 `json:"output_tokens"`
}

func AnthropicChatStream(apiKey string, voices []data.CharacterConfigVoice, multi bool, ttsOptimization bool, chatSystemPropmt string, temperature *float32, model string, cm []data.ChatCompletionMessage, text string, persistUserText bool, image *data.Image, chunkMessage chan api.ChunkMessage) ([]data.ChatCompletionMessage, *data.Tokens, error) {
	ctx := context.Background()

	body := buildAnthropicMessagesRequest(chatSystemPropmt, temperature, model, cm, text, image)

	reqBody, err := json.Marshal(body)
	if err != nil {
		return cm, nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, anthropicMessagesEndpoint, bytes.NewReader(reqBody))
	if err != nil {
		return cm, nil, err
	}
	req.Header.Set("content-type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", anthropicVersion)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return cm, nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return cm, nil, decodeAnthropicError(resp)
	}

	charChannel := make(chan rune)
	done := make(chan error)
	var t *data.Tokens

	defer close(charChannel)
	defer close(done)
	go func() {
		done <- streamAnthropicResponse(resp.Body, charChannel, &t)
	}()

	cr, err := chatReceiver(charChannel, done, multi, ttsOptimization, voices, chunkMessage, text, persistUserText, image, cm)
	return cr, t, err
}

func buildAnthropicMessagesRequest(chatSystemPropmt string, temperature *float32, model string, cm []data.ChatCompletionMessage, text string, image *data.Image) anthropicMessagesRequest {
	ncm := append(cm, data.ChatCompletionMessage{
		Role:    data.ChatCompletionMessageRoleUser,
		Content: text,
		Image:   image,
	})

	messages := make([]anthropicMessage, len(ncm))
	for i, v := range ncm {
		if v.Image == nil || i != len(ncm)-1 {
			messages[i] = anthropicMessage{
				Role:    v.Role,
				Content: v.Content,
			}
			continue
		}

		content := []any{
			anthropicImageBlock{
				Type: "image",
				Source: anthropicImageSource{
					Type:      "base64",
					MediaType: v.Image.MediaType(),
					Data:      v.Image.Base64(),
				},
			},
		}
		if v.Content != "" {
			content = append(content, anthropicTextBlock{
				Type: "text",
				Text: v.Content,
			})
		}
		messages[i] = anthropicMessage{
			Role:    v.Role,
			Content: content,
		}
	}

	req := anthropicMessagesRequest{
		Model:     model,
		Messages:  messages,
		MaxTokens: 4096,
		Stream:    true,
	}

	if chatSystemPropmt != "" {
		req.System = []anthropicTextBlock{
			{
				Type: "text",
				Text: chatSystemPropmt,
			},
		}
	}

	if temperature != nil {
		v := float64(*temperature)
		req.Temperature = &v
	}

	return req
}

func decodeAnthropicError(resp *http.Response) error {
	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return fmt.Errorf("anthropic API error: %s (failed to read body: %w)", resp.Status, readErr)
	}

	var apiErr anthropicErrorResponse
	if err := json.Unmarshal(body, &apiErr); err == nil && apiErr.Error.Message != "" {
		return fmt.Errorf("anthropic API error: %s: %s", resp.Status, apiErr.Error.Message)
	}

	return fmt.Errorf("anthropic API error: %s: %s", resp.Status, strings.TrimSpace(string(body)))
}

func streamAnthropicResponse(body io.Reader, charChannel chan rune, tokens **data.Tokens) error {
	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 1024), 1024*1024)

	var eventType string
	var dataLines []string

	dispatch := func() error {
		if eventType == "" && len(dataLines) == 0 {
			return nil
		}

		payload := strings.Join(dataLines, "\n")
		defer func() {
			eventType = ""
			dataLines = nil
		}()

		switch eventType {
		case "", "ping", "content_block_start", "content_block_stop":
			return nil
		case "message_start":
			var event anthropicStreamMessageStartEvent
			if err := json.Unmarshal([]byte(payload), &event); err != nil {
				return err
			}
			*tokens = &data.Tokens{
				InputTokens:  event.Message.Usage.InputTokens,
				OutputTokens: event.Message.Usage.OutputTokens,
			}
			return nil
		case "content_block_delta":
			var event anthropicStreamContentBlockDeltaEvent
			if err := json.Unmarshal([]byte(payload), &event); err != nil {
				return err
			}
			if event.Delta.Type != "text_delta" {
				return nil
			}
			for _, c := range event.Delta.Text {
				charChannel <- c
			}
			return nil
		case "message_delta":
			var event anthropicStreamMessageDeltaEvent
			if err := json.Unmarshal([]byte(payload), &event); err != nil {
				return err
			}
			if *tokens == nil {
				*tokens = &data.Tokens{}
			}
			if event.Usage.InputTokens > 0 {
				(*tokens).InputTokens = event.Usage.InputTokens
			}
			(*tokens).OutputTokens = event.Usage.OutputTokens
			return nil
		case "message_stop":
			return io.EOF
		case "error":
			var apiErr anthropicErrorResponse
			if err := json.Unmarshal([]byte(payload), &apiErr); err != nil {
				return err
			}
			if apiErr.Error.Message == "" {
				return errors.New("anthropic stream error")
			}
			return errors.New(apiErr.Error.Message)
		default:
			return nil
		}
	}

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			err := dispatch()
			if errors.Is(err, io.EOF) {
				return nil
			}
			if err != nil {
				return err
			}
			continue
		}

		if strings.HasPrefix(line, ":") {
			continue
		}
		if strings.HasPrefix(line, "event:") {
			eventType = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
			continue
		}
		if strings.HasPrefix(line, "data:") {
			dataLines = append(dataLines, strings.TrimSpace(strings.TrimPrefix(line, "data:")))
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	err := dispatch()
	if errors.Is(err, io.EOF) {
		return nil
	}
	return err
}
