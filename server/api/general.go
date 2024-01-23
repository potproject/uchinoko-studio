package api

import (
	"github.com/potproject/uchinoko-studio/envgen"
	openai "github.com/sashabaranov/go-openai"
)

type TextMessage struct {
	Text    string
	IsFirst bool
	IsFinal bool
}

type OpenAIClientExtend struct {
	Client *openai.Client
	ApiKey string
}

func OpenAINewClient() *OpenAIClientExtend {
	return &OpenAIClientExtend{
		Client: openai.NewClient(envgen.Get().OPENAI_API_KEY()),
		ApiKey: envgen.Get().OPENAI_API_KEY(),
	}
}
