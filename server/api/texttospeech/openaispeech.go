package texttospeech

import (
	"context"
	"io"

	"github.com/sashabaranov/go-openai"
)

func openAISpeech(apiKey string, model string, voice string, text string) ([]byte, error) {
	ctx := context.Background()
	client := openai.NewClient(apiKey)

	req := openai.CreateSpeechRequest{
		Model:          openai.SpeechModel(model),
		Voice:          openai.SpeechVoice(voice),
		ResponseFormat: "wav",
		Input:          text,
	}

	resp, err := client.CreateSpeech(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Close()

	return io.ReadAll(resp)
}
