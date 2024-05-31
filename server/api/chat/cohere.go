package chat

import (
	"context"
	"errors"
	"io"

	cohere "github.com/cohere-ai/cohere-go/v2"
	cohereclient "github.com/cohere-ai/cohere-go/v2/client"
	"github.com/potproject/uchinoko-studio/api"
	"github.com/potproject/uchinoko-studio/data"
)

func CohereChatStream(apiKey string, voices []data.CharacterConfigVoice, multi bool, chatSystemPropmt string, model string, cm []data.ChatCompletionMessage, text string, image *data.Image, chunkMessage chan api.ChunkMessage) ([]data.ChatCompletionMessage, *data.Tokens, error) {
	ctx := context.Background()
	// Tokens is Not Implemented
	var t *data.Tokens
	c := cohereclient.NewClient(cohereclient.WithToken(apiKey))

	cohereChatMessages := make([]*cohere.ChatMessage, len(cm)+1)
	cohereChatMessages[0] = &cohere.ChatMessage{
		Role:    "SYSTEM",
		Message: chatSystemPropmt,
	}
	for i, v := range cm {
		cohereRole := "USER"
		if v.Role == data.ChatCompletionMessageRoleAssistant {
			cohereRole = "CHATBOT"
		}
		cohereChatMessages[i+1] = &cohere.ChatMessage{
			Role:    cohere.ChatMessageRole(cohereRole),
			Message: v.Content,
		}
	}

	stream, err := c.ChatStream(
		ctx,
		&cohere.ChatStreamRequest{
			Message:     text,
			Model:       &model,
			ChatHistory: cohereChatMessages,
		},
	)
	if err != nil {
		return nil, t, err
	}

	defer stream.Close()

	charChannel := make(chan rune)
	done := make(chan error)

	defer close(charChannel)
	defer close(done)
	go func() {
		for {
			response, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				done <- nil
				break
			}
			if err != nil {
				done <- err
				break
			}
			if response.TextGeneration == nil {
				continue
			}
			content := response.TextGeneration.Text
			for _, c := range content {
				charChannel <- c
			}
		}
	}()

	cr, err := chatReceiver(charChannel, done, multi, voices, chunkMessage, text, image, cm)
	return cr, t, err
}
