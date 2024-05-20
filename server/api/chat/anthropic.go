package chat

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"

	claude "github.com/potproject/claude-sdk-go"
	"github.com/potproject/uchinoko-studio/api"
	"github.com/potproject/uchinoko-studio/data"
)

func AnthropicChatStream(apiKey string, voices []data.CharacterConfigVoice, multi bool, chatSystemPropmt string, model string, cm []data.ChatCompletionMessage, text string, image *data.Image, chunkMessage chan api.ChunkMessage) ([]data.ChatCompletionMessage, error) {
	ctx := context.Background()
	c := claude.NewClient(apiKey)
	ncm := append(cm, data.ChatCompletionMessage{
		Role:    data.ChatCompletionMessageRoleUser,
		Content: text,
		Image:   image,
	})

	anthropicChatMessages := make([]claude.RequestBodyMessagesMessages, len(ncm))
	for i, v := range ncm {
		if v.Image == nil {
			anthropicChatMessages[i] = claude.RequestBodyMessagesMessages{
				Role:    v.Role,
				Content: v.Content,
			}
		} else {
			anthropicChatMessages[i] = claude.RequestBodyMessagesMessages{
				Role: v.Role,
				ContentTypeImage: []claude.RequestBodyMessagesMessagesContentTypeImage{
					{
						Source: claude.RequestBodyMessagesMessagesContentTypeImageSource{
							Type:      "base64",
							MediaType: v.Image.MediaType(),
							Data:      v.Image.Base64(),
						},
					},
				},
			}
		}
	}

	body := claude.RequestBodyMessages{
		Model:     model,
		MaxTokens: 4096,
		Messages:  anthropicChatMessages,
		Stream:    true,
		System:    chatSystemPropmt,
	}

	stream, err := c.CreateMessagesStream(ctx, body)
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
			if response.Content == nil || len(response.Content) == 0 {
				continue
			}
			content := response.Content[0].Text
			for _, c := range content {
				charChannel <- c
			}
		}
		done <- true
	}()

	return chatReceiver(charChannel, done, multi, voices, chunkMessage, text, image, cm)
}
