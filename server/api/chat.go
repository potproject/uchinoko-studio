package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/potproject/uchinoko-studio/db"
	"github.com/potproject/uchinoko-studio/envgen"
	openai "github.com/sashabaranov/go-openai"
	"github.com/tmaxmax/go-sse"
)

const chars = ".,?!;:—-()[]{} 。、？！；：「」（）［］｛｝　\"'"

func OpenAIChatStream(apiKey string, cm []openai.ChatCompletionMessage, text string, chunkMessage chan TextMessage, responseText chan string) ([]openai.ChatCompletionMessage, error) {
	ctx := context.Background()
	c := openai.NewClient(apiKey)
	ncm := append(cm, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: text,
	})

	req := openai.ChatCompletionRequest{
		Model:     envgen.Get().CHAT_MODEL(),
		MaxTokens: 4096,
		Messages:  ncm,
		Stream:    true,
	}
	stream, err := c.CreateChatCompletionStream(ctx, req)
	if err != nil {
		log.Printf("ChatCompletionStream error: %v\n", err)
		return cm, err
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
			return append(
				ncm,
				openai.ChatCompletionMessage{
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
			return append(
				ncm,
				openai.ChatCompletionMessage{
					Role:    openai.ChatMessageRoleAssistant,
					Content: allText,
				},
			), nil
		}
	}
}

type AnthropicChatResponse struct {
	Type string `json:"type"`
}

const AntrhopicChatAPIEndpoint = "https://api.anthropic.com/v1/messages"

const (
	AnthropicChatResponseTypeMessageStart      = "message_start"
	AnthropicChatResponseTypeContentBlockStart = "content_block_start"
	AthropicChatResponseTypePing               = "ping"
	AnthropicChatResponseTypeContentBlockDelta = "content_block_delta"
	AnthropicChatResponseTypeContentBlockStop  = "content_block_stop"
	AnthropicChatResponseTypeMessageSDelta     = "message_delta"
	AnthropicChatResponseTypeMessageStop       = "message_stop"
	AnthropicChatResponseTypeError             = "error"
)

type AnthropicContentBlockDeltaBody struct {
	Type  string `json:"type"`
	Index int    `json:"index"`
	Delta struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"delta"`
}

type anthropicChatCompletionRequest struct {
	openai.ChatCompletionRequest
	System string `json:"system"`
}

func AnthropicChatStream(apiKey string, cm []openai.ChatCompletionMessage, text string, chunkMessage chan TextMessage, responseText chan string) ([]openai.ChatCompletionMessage, error) {
	ncm := append(cm, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: text,
	})

	body := anthropicChatCompletionRequest{
		ChatCompletionRequest: openai.ChatCompletionRequest{

			Model:     envgen.Get().CHAT_MODEL(),
			MaxTokens: 4096,
			Messages:  ncm,
			Stream:    true,
		},
		System: db.SystemMessage,
	}

	bodyJson, err := json.Marshal(body)
	if err != nil {
		return cm, err
	}

	client := sse.Client{
		Backoff: sse.Backoff{
			MaxRetries: -1,
		},
	}
	req, err := http.NewRequest(http.MethodPost, AntrhopicChatAPIEndpoint, strings.NewReader(string(bodyJson)))
	if err != nil {
		return cm, err
	}
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("content-type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	conn := client.NewConnection(req)

	allText := ""
	bufferText := ""
	firstSend := true
	unsubscribe := conn.SubscribeEvent(AnthropicChatResponseTypeContentBlockDelta, func(event sse.Event) {
		var body AnthropicContentBlockDeltaBody
		if err := json.Unmarshal([]byte(event.Data), &body); err != nil {
			fmt.Println(err)
			return
		}
		content := body.Delta.Text
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
	})
	if err := conn.Connect(); !errors.Is(err, io.EOF) {
		return cm, err
	}
	defer unsubscribe()
	responseText <- allText
	chunkMessage <- TextMessage{
		Text:    bufferText,
		IsFirst: firstSend,
		IsFinal: true,
	}
	return append(
		ncm,
		openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleAssistant,
			Content: allText,
		},
	), nil
}
