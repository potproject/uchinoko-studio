package api

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"

	cohere "github.com/cohere-ai/cohere-go/v2"
	cohereclient "github.com/cohere-ai/cohere-go/v2/client"
	claude "github.com/potproject/claude-sdk-go"
	"github.com/potproject/uchinoko-studio/data"
	openai "github.com/sashabaranov/go-openai"
)

const chars = ".,?!;:—-)]} 。、？！；：」）］｝　\"'"

type ChatStream func(string, []data.CharacterConfigVoice, bool, string, string, []data.ChatCompletionMessage, string, chan TextMessage) ([]data.ChatCompletionMessage, error)

func OpenAIChatStream(apiKey string, voices []data.CharacterConfigVoice, multi bool, chatSystemPropmt string, model string, cm []data.ChatCompletionMessage, text string, chunkMessage chan TextMessage) ([]data.ChatCompletionMessage, error) {
	ctx := context.Background()
	c := openai.NewClient(apiKey)

	voice := voices[0]
	voiceIndentifications := make([]string, len(voices))
	if multi {
		for i, v := range voices {
			voiceIndentifications[i] = v.Identification
		}
	}

	ncm := append(cm, data.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
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

	allText := ""
	bufferText := ""
	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			chunkMessage <- TextMessage{
				Text:  bufferText,
				Voice: voice,
			}
			return append(
				ncm,
				data.ChatCompletionMessage{
					Role:    openai.ChatMessageRoleAssistant,
					Content: allText,
				},
			), nil
		}

		if err != nil {
			fmt.Printf("\nStream error: %v\n", err)
			return cm, err
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
				Text:  bufferText + content,
				Voice: voice,
			}
			bufferText = ""
		} else {
			bufferText += content
		}

		// bufferTextに voiceIndentifications が含まれている場合
		if multi {
			for i, v := range voiceIndentifications {
				if strings.Contains(bufferText, v) {
					bufferText = strings.Replace(bufferText, v, "", -1)
					chunkMessage <- TextMessage{
						Text:  bufferText,
						Voice: voice,
					}
					bufferText = ""
					voice = voices[i]
					break
				}
			}
		}

		if response.Choices[0].FinishReason == "stop" {
			chunkMessage <- TextMessage{
				Text:  bufferText,
				Voice: voice,
			}
			return append(
				ncm,
				data.ChatCompletionMessage{
					Role:    openai.ChatMessageRoleAssistant,
					Content: allText,
				},
			), nil
		}
	}
}

func AnthropicChatStream(apiKey string, voices []data.CharacterConfigVoice, multi bool, chatSystemPropmt string, model string, cm []data.ChatCompletionMessage, text string, chunkMessage chan TextMessage) ([]data.ChatCompletionMessage, error) {
	ctx := context.Background()
	c := claude.NewClient(apiKey)
	ncm := append(cm, data.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: text,
	})

	voice := voices[0]
	voiceIndentifications := make([]string, len(voices))
	if multi {
		for i, v := range voices {
			voiceIndentifications[i] = v.Identification
		}
	}

	anthropicChatMessages := make([]claude.RequestBodyMessagesMessages, len(ncm))
	for i, v := range ncm {
		anthropicChatMessages[i] = claude.RequestBodyMessagesMessages{
			Role:    v.Role,
			Content: v.Content,
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

	allText := ""
	bufferText := ""
	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			chunkMessage <- TextMessage{
				Text:  bufferText,
				Voice: voice,
			}
			return append(
				ncm,
				data.ChatCompletionMessage{
					Role:    openai.ChatMessageRoleAssistant,
					Content: allText,
				},
			), nil
		}

		if err != nil {
			fmt.Printf("\nStream error: %v\n", err)
			return cm, err
		}
		content := response.Content[0].Text
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
				Text:  bufferText + content,
				Voice: voice,
			}
			bufferText = ""
		} else {
			bufferText += content
		}

		// bufferTextに voiceIndentifications が含まれている場合
		if multi {
			for i, v := range voiceIndentifications {
				if strings.Contains(bufferText, v) {
					bufferText = strings.Replace(bufferText, v, "", -1)
					chunkMessage <- TextMessage{
						Text:  bufferText,
						Voice: voice,
					}
					bufferText = ""
					voice = voices[i]
					break
				}
			}
		}
	}
}

func CohereChatStream(apiKey string, voices []data.CharacterConfigVoice, multi bool, chatSystemPropmt string, model string, cm []data.ChatCompletionMessage, text string, chunkMessage chan TextMessage) ([]data.ChatCompletionMessage, error) {
	ctx := context.Background()
	c := cohereclient.NewClient(cohereclient.WithToken(apiKey))

	voice := voices[0]
	voiceIndentifications := make([]string, len(voices))
	if multi {
		for i, v := range voices {
			voiceIndentifications[i] = v.Identification
		}
	}

	cohereChatMessages := make([]*cohere.ChatMessage, len(cm)+1)
	cohereChatMessages[0] = &cohere.ChatMessage{
		Role:    "SYSTEM",
		Message: chatSystemPropmt,
	}
	for i, v := range cm {
		cohereRole := "USER"
		if v.Role == openai.ChatMessageRoleAssistant {
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

	allText := ""
	bufferText := ""
	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			chunkMessage <- TextMessage{
				Text:  bufferText,
				Voice: voice,
			}
			return append(
				cm,
				data.ChatCompletionMessage{
					Role:    openai.ChatMessageRoleUser,
					Content: text,
				},
				data.ChatCompletionMessage{
					Role:    openai.ChatMessageRoleAssistant,
					Content: allText,
				},
			), nil
		}
		if err != nil {
			fmt.Printf("\nStream error: %v\n", err)
			return cm, err
		}
		if response.TextGeneration == nil {
			continue
		}
		content := response.TextGeneration.Text
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
				Text:  bufferText + content,
				Voice: voice,
			}
			bufferText = ""
		} else {
			bufferText += content
		}

		// bufferTextに voiceIndentifications が含まれている場合
		if multi {
			for i, v := range voiceIndentifications {
				if strings.Contains(bufferText, v) {
					bufferText = strings.Replace(bufferText, v, "", -1)
					chunkMessage <- TextMessage{
						Text:  bufferText,
						Voice: voice,
					}
					bufferText = ""
					voice = voices[i]
					break
				}
			}
		}
	}
}
