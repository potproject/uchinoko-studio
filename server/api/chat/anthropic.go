package chat

import (
	"context"
	"errors"
	"io"
	"log"

	claude "github.com/potproject/claude-sdk-go"
	"github.com/potproject/uchinoko-studio/api"
	"github.com/potproject/uchinoko-studio/data"
)

func AnthropicChatStream(apiKey string, voices []data.CharacterConfigVoice, multi bool, chatSystemPropmt string, model string, cm []data.ChatCompletionMessage, text string, image *data.Image, chunkMessage chan api.ChunkMessage) ([]data.ChatCompletionMessage, *data.Tokens, error) {
	ctx := context.Background()
	c := claude.NewClient(apiKey)

	var t *data.Tokens
	ncm := append(cm, data.ChatCompletionMessage{
		Role:    data.ChatCompletionMessageRoleUser,
		Content: text,
		Image:   image,
	})

	anthropicChatMessages := make([]claude.RequestBodyMessagesMessages, len(ncm))
	for i, v := range ncm {
		if v.Image == nil || i != len(ncm)-1 {
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
			t = &data.Tokens{
				InputTokens:  response.Usage.InputTokens,
				OutputTokens: response.Usage.OutputTokens,
			}
			if errors.Is(err, io.EOF) {
				done <- nil
				break
			}
			if err != nil {
				done <- err
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
	}()
	cr, err := chatReceiver(charChannel, done, multi, voices, chunkMessage, text, image, cm)
	return cr, t, err
}
