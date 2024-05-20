package chat

import (
	"context"
	"errors"
	"fmt"
	"io"

	cohere "github.com/cohere-ai/cohere-go/v2"
	cohereclient "github.com/cohere-ai/cohere-go/v2/client"
	"github.com/potproject/uchinoko-studio/api"
	"github.com/potproject/uchinoko-studio/data"
)

func CohereChatStream(apiKey string, voices []data.CharacterConfigVoice, multi bool, chatSystemPropmt string, model string, cm []data.ChatCompletionMessage, text string, image *data.Image, chunkMessage chan api.ChunkMessage) ([]data.ChatCompletionMessage, error) {
	ctx := context.Background()
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
		return nil, err
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
			if response.TextGeneration == nil {
				continue
			}
			content := response.TextGeneration.Text
			for _, c := range content {
				charChannel <- c
			}
		}
		done <- true
	}()

	return chatReceiver(charChannel, done, multi, voices, chunkMessage, text, image, cm)
}
