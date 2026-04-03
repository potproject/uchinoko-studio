package memory

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	gemini "github.com/google/generative-ai-go/genai"
	"github.com/potproject/uchinoko-studio/data"
	"github.com/potproject/uchinoko-studio/envgen"
	openai "github.com/sashabaranov/go-openai"
	"google.golang.org/api/option"
)

func completeText(character data.CharacterConfig, systemPrompt string, userPrompt string) (string, error) {
	switch character.Chat.Type {
	case "openai":
		return completeOpenAI(envgen.Get().OPENAI_API_KEY(), "", character.Chat.Model, systemPrompt, userPrompt)
	case "openai-local":
		return completeOpenAI(envgen.Get().OPENAI_LOCAL_API_KEY(), envgen.Get().OPENAI_LOCAL_API_ENDPOINT(), character.Chat.Model, systemPrompt, userPrompt)
	case "deepseek":
		return completeOpenAICompatible("https://api.deepseek.com", envgen.Get().DEEPSEEK_API_KEY(), character.Chat.Model, systemPrompt, userPrompt)
	case "anthropic":
		return completeAnthropic(envgen.Get().ANTHROPIC_API_KEY(), character.Chat.Model, systemPrompt, userPrompt)
	case "gemini":
		return completeGemini(envgen.Get().GEMINI_API_KEY(), character.Chat.Model, systemPrompt, userPrompt)
	default:
		return "", fmt.Errorf("unsupported memory model provider: %s", character.Chat.Type)
	}
}

func completeOpenAI(apiKey string, endpoint string, model string, systemPrompt string, userPrompt string) (string, error) {
	if endpoint == "" {
		client := openai.NewClient(apiKey)
		return completeOpenAIWithClient(client, model, systemPrompt, userPrompt)
	}
	config := openai.DefaultConfig(apiKey)
	baseURL, _ := url.JoinPath(endpoint, "v1")
	config.BaseURL = baseURL
	client := openai.NewClientWithConfig(config)
	return completeOpenAIWithClient(client, model, systemPrompt, userPrompt)
}

func completeOpenAICompatible(endpoint string, apiKey string, model string, systemPrompt string, userPrompt string) (string, error) {
	config := openai.DefaultConfig(apiKey)
	config.BaseURL = endpoint
	client := openai.NewClientWithConfig(config)
	return completeOpenAIWithClient(client, model, systemPrompt, userPrompt)
}

func completeOpenAIWithClient(client *openai.Client, model string, systemPrompt string, userPrompt string) (string, error) {
	resp, err := client.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
		Model: model,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: systemPrompt},
			{Role: openai.ChatMessageRoleUser, Content: userPrompt},
		},
		Temperature: 0,
	})
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", nil
	}
	return strings.TrimSpace(resp.Choices[0].Message.Content), nil
}

func completeAnthropic(apiKey string, model string, systemPrompt string, userPrompt string) (string, error) {
	body := map[string]any{
		"model":      model,
		"max_tokens": 1200,
		"system": []map[string]string{
			{"type": "text", "text": systemPrompt},
		},
		"messages": []map[string]any{
			{
				"role":    "user",
				"content": []map[string]string{{"type": "text", "text": userPrompt}},
			},
		},
		"temperature": 0,
	}
	encoded, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, "https://api.anthropic.com/v1/messages", bytes.NewReader(encoded))
	if err != nil {
		return "", err
	}
	req.Header.Set("content-type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("anthropic API error: %s", strings.TrimSpace(string(bodyBytes)))
	}
	var decoded struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(bodyBytes, &decoded); err != nil {
		return "", err
	}
	if len(decoded.Content) == 0 {
		return "", nil
	}
	return strings.TrimSpace(decoded.Content[0].Text), nil
}

func completeGemini(apiKey string, model string, systemPrompt string, userPrompt string) (string, error) {
	client, err := gemini.NewClient(context.Background(), option.WithAPIKey(apiKey))
	if err != nil {
		return "", err
	}
	defer client.Close()
	generativeModel := client.GenerativeModel(model)
	generativeModel.Temperature = float32Ptr(0)
	generativeModel.SystemInstruction = &gemini.Content{
		Parts: []gemini.Part{gemini.Text(systemPrompt)},
	}
	resp, err := generativeModel.GenerateContent(context.Background(), gemini.Text(userPrompt))
	if err != nil {
		return "", err
	}
	if resp == nil || len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
		return "", nil
	}
	var out strings.Builder
	for _, part := range resp.Candidates[0].Content.Parts {
		out.WriteString(fmt.Sprintf("%s", part))
	}
	return strings.TrimSpace(out.String()), nil
}

func float32Ptr(v float32) *float32 {
	return &v
}
