package api

import (
	"context"

	claude "github.com/potproject/claude-sdk-go"
	openai "github.com/sashabaranov/go-openai"
)

func OpenAIChat(apiKey string, chatSystemPropmt string, model string, text string) (string, error) {
	ctx := context.Background()
	c := openai.NewClient(apiKey)

	req := openai.ChatCompletionRequest{
		Model:     model,
		MaxTokens: 4096,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: chatSystemPropmt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: text,
			},
		},
	}
	resp, err := c.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", err
	}
	return resp.Choices[0].Message.Content, nil
}

func AnthropicChat(apiKey string, chatSystemPropmt string, model string, text string) (string, error) {
	ctx := context.Background()
	c := claude.NewClient(apiKey)

	m := claude.RequestBodyMessages{
		Model:     model,
		MaxTokens: 4096,
		Messages: []claude.RequestBodyMessagesMessages{
			{
				Role:    claude.MessagesRoleUser,
				Content: text,
			},
		},
		System: chatSystemPropmt,
	}

	res, err := c.CreateMessages(ctx, m)
	if err != nil {
		return "", err
	}
	return res.Content[0].Text, nil
}
