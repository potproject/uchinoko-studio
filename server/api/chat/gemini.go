package chat

import (
	"context"
	"fmt"
	"strings"

	"github.com/potproject/uchinoko-studio/api"
	"github.com/potproject/uchinoko-studio/data"

	gemini "github.com/google/generative-ai-go/genai"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

func GeminiChatStream(apiKey string, voices []data.CharacterConfigVoice, multi bool, chatSystemPropmt string, model string, cm []data.ChatCompletionMessage, text string, image *data.Image, chunkMessage chan api.ChunkMessage) ([]data.ChatCompletionMessage, error) {
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

	geminiContents := make([]*gemini.Content, len(cm))
	for i, v := range cm {
		geminiRole := "user"
		if v.Role == data.ChatCompletionMessageRoleAssistant {
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

	var part gemini.Part
	if image == nil {
		part = gemini.Text(text)
	} else {
		part = gemini.Blob{
			MIMEType: image.MediaType(),
			Data:     image.Data,
		}
	}
	iter := cs.SendMessageStream(
		ctx,
		part,
	)

	charChannel := make(chan rune)
	done := make(chan bool)

	defer close(charChannel)
	defer close(done)
	go func() {
		for {
			response, err := iter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				fmt.Printf("\nStream error: %v\n", err)
				break
			}
			gres := geminiResponse(response)
			if gres == nil {
				continue
			}
			content := *gres
			for _, c := range content {
				charChannel <- c
			}
		}
		done <- true
	}()
	return chatReceiver(charChannel, done, multi, voices, chunkMessage, text, image, cm)
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
