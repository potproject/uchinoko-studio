package chat

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/url"

	"github.com/potproject/uchinoko-studio/api"
	"github.com/potproject/uchinoko-studio/data"
	"github.com/potproject/uchinoko-studio/envgen"
	openai "github.com/sashabaranov/go-openai"
)

func OpenAILocalChatStream(apiKey string, voices []data.CharacterConfigVoice, multi bool, chatSystemPropmt string, model string, cm []data.ChatCompletionMessage, text string, chunkMessage chan api.ChunkMessage) ([]data.ChatCompletionMessage, error) {
	config := openai.DefaultConfig(apiKey)
	baseUrl, _ := url.JoinPath(envgen.Get().OPENAI_LOCAL_API_ENDPOINT(), "v1")
	config.BaseURL = baseUrl
	c := openai.NewClientWithConfig(config)
	return OpenAIChatStreamMain(context.Background(), c, voices, multi, chatSystemPropmt, model, cm, text, chunkMessage)
}

func OpenAIChatStream(apiKey string, voices []data.CharacterConfigVoice, multi bool, chatSystemPropmt string, model string, cm []data.ChatCompletionMessage, text string, chunkMessage chan api.ChunkMessage) ([]data.ChatCompletionMessage, error) {
	ctx := context.Background()
	c := openai.NewClient(apiKey)
	return OpenAIChatStreamMain(ctx, c, voices, multi, chatSystemPropmt, model, cm, text, chunkMessage)
}

func OpenAIChatStreamMain(ctx context.Context, c *openai.Client, voices []data.CharacterConfigVoice, multi bool, chatSystemPropmt string, model string, cm []data.ChatCompletionMessage, text string, chunkMessage chan api.ChunkMessage) ([]data.ChatCompletionMessage, error) {
	ncm := append(cm, data.ChatCompletionMessage{
		Role:    data.ChatCompletionMessageRoleUser,
		Content: text,
	})
	openaiChatMessages := make([]openai.ChatCompletionMessage, len(ncm))
	for i, v := range ncm {
		openaiChatMessages[i] = openai.ChatCompletionMessage{
			Role:    v.Role,
			Content: v.Content,
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
	}
	stream, err := c.CreateChatCompletionStream(ctx, req)
	if err != nil {
		log.Printf("ChatCompletionStream error: %v\n", err)
		return cm, err
	}
	defer stream.Close()

	charChannel := make(chan rune)
	done := make(chan bool)

	defer close(charChannel)
	defer close(done)
	go func() {
		for {
			response, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				fmt.Printf("\nStream error: %v\n", err)
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
		done <- true
	}()

	return chatReceiver(charChannel, done, multi, voices, chunkMessage, text, cm)
}
