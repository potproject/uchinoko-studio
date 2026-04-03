package memory

import (
	"context"
	"net/url"
	"strings"

	"github.com/potproject/uchinoko-studio/data"
	"github.com/potproject/uchinoko-studio/envgen"
	openai "github.com/sashabaranov/go-openai"
)

func embedText(text string, cfg data.CharacterConfigMemory) ([]float64, error) {
	text = NormalizeText(text)
	if text == "" {
		return nil, nil
	}
	cfg = memoryDefaults(cfg)
	if apiKey := strings.TrimSpace(envgen.Get().OPENAI_API_KEY()); apiKey != "" {
		client := openai.NewClient(apiKey)
		return createEmbedding(client, cfg.EmbeddingModel, text)
	}
	if endpoint := strings.TrimSpace(envgen.Get().OPENAI_LOCAL_API_ENDPOINT()); endpoint != "" && strings.TrimSpace(cfg.EmbeddingModel) != "" {
		config := openai.DefaultConfig(envgen.Get().OPENAI_LOCAL_API_KEY())
		baseURL, _ := url.JoinPath(endpoint, "v1")
		config.BaseURL = baseURL
		client := openai.NewClientWithConfig(config)
		return createEmbedding(client, cfg.EmbeddingModel, text)
	}
	return nil, nil
}

func createEmbedding(client *openai.Client, model string, text string) ([]float64, error) {
	resp, err := client.CreateEmbeddings(context.Background(), openai.EmbeddingRequest{
		Model: openai.EmbeddingModel(model),
		Input: []string{text},
	})
	if err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, nil
	}
	out := make([]float64, len(resp.Data[0].Embedding))
	for i, value := range resp.Data[0].Embedding {
		out[i] = float64(value)
	}
	return out, nil
}
