package api

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/potproject/uchinoko/envgen"
	openai "github.com/sashabaranov/go-openai"
)

const chars = ".,?!;:—-()[]{} 。、？！；：ー「」（）［］｛｝　\"'"

func ChatStream(c *OpenAIClientExtend, text string, chunkMessage chan<- TextMessage, responseText chan string) error {
	ctx := context.Background()

	req := openai.ChatCompletionRequest{
		Model:     envgen.Get().OPENAI_CHAT_MODEL(),
		MaxTokens: 4096,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: text,
			},
		},
		Stream: true,
	}
	stream, err := c.Client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		log.Printf("ChatCompletionStream error: %v\n", err)
		return err
	}
	defer stream.Close()

	allText := ""
	bufferText := ""
	firstSend := true
	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			responseText <- allText
			chunkMessage <- TextMessage{
				Text:    bufferText,
				IsFirst: firstSend,
				IsFinal: true,
			}
			return nil
		}

		if err != nil {
			fmt.Printf("\nStream error: %v\n", err)
			return err
		}
		content := response.Choices[0].Delta.Content
		allText += content
		chunked := false
		for _, c := range content {
			chunked = strings.Contains(chars, string(c))
			if chunked {
				break
			}
		}
		if chunked {
			chunkMessage <- TextMessage{
				Text:    bufferText + content,
				IsFirst: firstSend,
				IsFinal: false,
			}
			firstSend = false
			bufferText = ""
		} else {
			bufferText += content
		}

		if response.Choices[0].FinishReason == "stop" {
			responseText <- allText
			chunkMessage <- TextMessage{
				Text:    bufferText,
				IsFirst: firstSend,
				IsFinal: true,
			}
			return nil
		}
	}
}
