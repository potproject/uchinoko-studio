package chat

import (
	"context"

	"github.com/potproject/uchinoko-studio/api"
	"github.com/potproject/uchinoko-studio/data"
	openai "github.com/sashabaranov/go-openai"
)

func DeepSeekChatStream(apiKey string, voices []data.CharacterConfigVoice, multi bool, ttsOptimization bool, chatSystemPropmt string, model string, cm []data.ChatCompletionMessage, text string, image *data.Image, chunkMessage chan api.ChunkMessage) ([]data.ChatCompletionMessage, *data.Tokens, error) {
	baseUrl := "https://api.deepseek.com/v1"
	config := openai.DefaultConfig(apiKey)
	config.BaseURL = baseUrl
	c := openai.NewClientWithConfig(config)
	return OpenAIChatStreamMain(context.Background(), c, voices, multi, ttsOptimization, chatSystemPropmt, model, cm, text, image, chunkMessage)
}
