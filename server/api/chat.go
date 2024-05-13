package api

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/url"
	"strings"

	cohere "github.com/cohere-ai/cohere-go/v2"
	cohereclient "github.com/cohere-ai/cohere-go/v2/client"
	gemini "github.com/google/generative-ai-go/genai"
	claude "github.com/potproject/claude-sdk-go"
	"github.com/potproject/uchinoko-studio/data"
	"github.com/potproject/uchinoko-studio/envgen"
	openai "github.com/sashabaranov/go-openai"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

const chars = ".,?!;:—-)]} 。、？！；：」）］｝　\"'"

type ChatStream func(string, []data.CharacterConfigVoice, bool, string, string, []data.ChatCompletionMessage, string, chan TextMessage) ([]data.ChatCompletionMessage, error)

func OpenAILocalChatStream(apiKey string, voices []data.CharacterConfigVoice, multi bool, chatSystemPropmt string, model string, cm []data.ChatCompletionMessage, text string, chunkMessage chan TextMessage) ([]data.ChatCompletionMessage, error) {
	config := openai.DefaultConfig(apiKey)
	baseUrl, _ := url.JoinPath(envgen.Get().OPENAI_LOCAL_API_ENDPOINT(), "v1")
	config.BaseURL = baseUrl
	c := openai.NewClientWithConfig(config)
	return OpenAIChatStreamMain(context.Background(), c, voices, multi, chatSystemPropmt, model, cm, text, chunkMessage)
}

func OpenAIChatStream(apiKey string, voices []data.CharacterConfigVoice, multi bool, chatSystemPropmt string, model string, cm []data.ChatCompletionMessage, text string, chunkMessage chan TextMessage) ([]data.ChatCompletionMessage, error) {
	ctx := context.Background()
	c := openai.NewClient(apiKey)
	return OpenAIChatStreamMain(ctx, c, voices, multi, chatSystemPropmt, model, cm, text, chunkMessage)
}

func OpenAIChatStreamMain(ctx context.Context, c *openai.Client, voices []data.CharacterConfigVoice, multi bool, chatSystemPropmt string, model string, cm []data.ChatCompletionMessage, text string, chunkMessage chan TextMessage) ([]data.ChatCompletionMessage, error) {
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

func GeminiChatStream(apiKey string, voices []data.CharacterConfigVoice, multi bool, chatSystemPropmt string, model string, cm []data.ChatCompletionMessage, text string, chunkMessage chan TextMessage) ([]data.ChatCompletionMessage, error) {
	ctx := context.Background()
	client, err := gemini.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, err
	}
	defer client.Close()

	geminiModel := client.GenerativeModel(model)
	geminiModel.SystemInstruction = &gemini.Content{
		Parts: []gemini.Part{gemini.Text(chatSystemPropmt)},
	}
	cs := geminiModel.StartChat()

	voice := voices[0]
	voiceIndentifications := make([]string, len(voices))
	if multi {
		for i, v := range voices {
			voiceIndentifications[i] = v.Identification
		}
	}

	geminiContents := make([]*gemini.Content, len(cm))
	for i, v := range cm {
		geminiRole := "user"
		if v.Role == openai.ChatMessageRoleAssistant {
			geminiRole = "model"
		}
		geminiContents[i] = &gemini.Content{
			Parts: []gemini.Part{
				gemini.Text(v.Content),
			},
			Role: geminiRole,
		}
	}
	cs.History = geminiContents

	iter := cs.SendMessageStream(
		ctx,
		gemini.Text(text),
	)

	allText := ""
	bufferText := ""
	for {
		response, err := iter.Next()
		if err == iterator.Done {
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
					Content: strings.Trim(allText, "\n"),
				},
			), nil
		}
		if err != nil {
			fmt.Printf("\nStream error: %v\n", err)
			return cm, err
		}
		gres := geminiResponse(response)
		if gres == nil {
			continue
		}
		content := *gres
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

func geminiResponse(resp *gemini.GenerateContentResponse) *string {
	var content string
	if resp == nil || resp.Candidates == nil || len(resp.Candidates) == 0 {
		return &content
	}
	for _, cand := range resp.Candidates {
		if cand.Content != nil {
			for _, part := range cand.Content.Parts {
				content += fmt.Sprintf("%s", part)
			}
		}
	}
	if len(content) == 0 {
		return nil
	}
	content = strings.Trim(content, "\n")
	return &content
}
