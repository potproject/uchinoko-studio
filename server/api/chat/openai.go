package chat

import (
	"context"
	"errors"
	"io"
	"log"
	"net/url"

	"github.com/potproject/uchinoko-studio/api"
	"github.com/potproject/uchinoko-studio/data"
	"github.com/potproject/uchinoko-studio/envgen"
	openai "github.com/sashabaranov/go-openai"
)

func OpenAILocalChatStream(apiKey string, voices []data.CharacterConfigVoice, multi bool, ttsOptimization bool, chatSystemPropmt string, temperature *float32, model string, cm []data.ChatCompletionMessage, text string, image *data.Image, chunkMessage chan api.ChunkMessage) ([]data.ChatCompletionMessage, *data.Tokens, error) {
	config := openai.DefaultConfig(apiKey)
	baseUrl, _ := url.JoinPath(envgen.Get().OPENAI_LOCAL_API_ENDPOINT(), "v1")
	config.BaseURL = baseUrl
	c := openai.NewClientWithConfig(config)
	return OpenAIChatStreamMain(context.Background(), c, voices, multi, ttsOptimization, chatSystemPropmt, temperature, model, cm, text, image, chunkMessage)
}

func OpenAIChatStream(apiKey string, voices []data.CharacterConfigVoice, multi bool, ttsOptimization bool, chatSystemPropmt string, temperature *float32, model string, cm []data.ChatCompletionMessage, text string, image *data.Image, chunkMessage chan api.ChunkMessage) ([]data.ChatCompletionMessage, *data.Tokens, error) {
	ctx := context.Background()
	c := openai.NewClient(apiKey)
	return OpenAIChatStreamMain(ctx, c, voices, multi, ttsOptimization, chatSystemPropmt, temperature, model, cm, text, image, chunkMessage)
}

func OpenAIChatStreamMain(ctx context.Context, c *openai.Client, voices []data.CharacterConfigVoice, multi bool, ttsOptimization bool, chatSystemPropmt string, temperature *float32, model string, cm []data.ChatCompletionMessage, text string, image *data.Image, chunkMessage chan api.ChunkMessage) ([]data.ChatCompletionMessage, *data.Tokens, error) {
	var t *data.Tokens
	ncm := append(cm, data.ChatCompletionMessage{
		Role:    data.ChatCompletionMessageRoleUser,
		Content: text,
		Image:   image,
	})
	openaiChatMessages := make([]openai.ChatCompletionMessage, len(ncm))
	for i, v := range ncm {
		if v.Image == nil || i != len(ncm)-1 {
			openaiChatMessages[i] = openai.ChatCompletionMessage{
				Role:    v.Role,
				Content: v.Content,
			}
		} else {
			openaiChatMessages[i] = openai.ChatCompletionMessage{
				Role: v.Role,
				MultiContent: []openai.ChatMessagePart{
					{
						Type: openai.ChatMessagePartTypeImageURL,
						ImageURL: &openai.ChatMessageImageURL{
							URL:    v.Image.DataURI(),
							Detail: openai.ImageURLDetailAuto,
						},
					},
				},
			}
			if v.Content != "" {
				openaiChatMessages[i].MultiContent = append(openaiChatMessages[i].MultiContent, openai.ChatMessagePart{
					Type: openai.ChatMessagePartTypeText,
					Text: v.Content,
				})
			}
		}
	}

	req := openai.ChatCompletionRequest{
		Model:     model,
		MaxTokens: 4096,
		Messages: append([]openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: chatSystemPropmt,
			},
		}, openaiChatMessages...),
		Stream: true,
		StreamOptions: &openai.StreamOptions{
			IncludeUsage: true,
		},
	}
	if temperature != nil {
		req.Temperature = *temperature
	}

	stream, err := c.CreateChatCompletionStream(ctx, req)
	if err != nil {
		log.Printf("ChatCompletionStream error: %v\n", err)
		return cm, t, err
	}
	defer stream.Close()

	charChannel := make(chan rune)
	done := make(chan error)

	defer close(charChannel)
	defer close(done)
	go func() {
		for {
			response, err := stream.Recv()
			if response.Usage != nil {
				t = &data.Tokens{
					InputTokens:  int64(response.Usage.PromptTokens),
					OutputTokens: int64(response.Usage.CompletionTokens),
				}
			}
			if errors.Is(err, io.EOF) {
				done <- nil
				break
			}
			if err != nil {
				done <- err
				break
			}
			if response.Choices == nil || len(response.Choices) == 0 {
				continue
			}
			content := response.Choices[0].Delta.Content
			for _, c := range content {
				charChannel <- c
			}
		}
	}()

	cr, err := chatReceiver(charChannel, done, multi, ttsOptimization, voices, chunkMessage, text, image, cm)
	return cr, t, err
}
