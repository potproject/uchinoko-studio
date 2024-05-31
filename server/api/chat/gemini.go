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

func setGeminiSafetySettings() []*gemini.SafetySetting {
	safeties := make([]*gemini.SafetySetting, 4)
	safeties[0] = &gemini.SafetySetting{
		Category:  gemini.HarmCategoryHarassment,
		Threshold: gemini.HarmBlockNone,
	}
	safeties[1] = &gemini.SafetySetting{
		Category:  gemini.HarmCategoryHateSpeech,
		Threshold: gemini.HarmBlockNone,
	}
	safeties[2] = &gemini.SafetySetting{
		Category:  gemini.HarmCategorySexuallyExplicit,
		Threshold: gemini.HarmBlockNone,
	}
	safeties[3] = &gemini.SafetySetting{
		Category:  gemini.HarmCategoryDangerousContent,
		Threshold: gemini.HarmBlockNone,
	}
	return safeties

}

func GeminiChatStream(apiKey string, voices []data.CharacterConfigVoice, multi bool, chatSystemPropmt string, model string, cm []data.ChatCompletionMessage, text string, image *data.Image, chunkMessage chan api.ChunkMessage) ([]data.ChatCompletionMessage, *data.Tokens, error) {
	ctx := context.Background()
	var t *data.Tokens
	client, err := gemini.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, t, err
	}
	defer client.Close()

	geminiModel := client.GenerativeModel(model)
	geminiModel.SafetySettings = setGeminiSafetySettings()
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
	done := make(chan error)

	defer close(charChannel)
	defer close(done)
	go func() {
		for {
			response, err := iter.Next()
			if response != nil && response.UsageMetadata != nil {
				t = &data.Tokens{
					InputTokens:  int64(response.UsageMetadata.PromptTokenCount),
					OutputTokens: int64(response.UsageMetadata.CandidatesTokenCount),
				}
			}
			if err == iterator.Done {
				done <- nil
				break
			}
			if err != nil {
				done <- err
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
	}()

	cr, err := chatReceiver(charChannel, done, multi, voices, chunkMessage, text, image, cm)
	return cr, t, err
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
