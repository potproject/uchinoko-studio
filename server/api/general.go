package api

import (
	"context"
	"time"

	"github.com/haguro/elevenlabs-go"
	"github.com/potproject/uchinoko-studio/envgen"
	openai "github.com/sashabaranov/go-openai"
)

const elevenlabsModelID = "eleven_multilingual_v2"

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

type ElevenLabsClientExtend struct {
	Client *elevenlabs.Client
	ApiKey string
}

func ElevenLabsNewClient() *ElevenLabsClientExtend {
	return &ElevenLabsClientExtend{
		Client: elevenlabs.NewClient(context.Background(), envgen.Get().ELEVENLABS_API_KEY(), 30*time.Second),
		ApiKey: envgen.Get().ELEVENLABS_API_KEY(),
	}
}
