package data

import "github.com/sashabaranov/go-openai"

type GeneralConfig struct {
	Transcription struct {
		Type string `json:"type"`
	} `json:"transcription"`
}

type CharacterConfigList struct {
	Characters []CharacterConfig `json:"characters"`
}

type CharacterConfig struct {
	General struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Image string `json:"image"`
	} `json:"general"`
	Voice struct {
		Type      string `json:"type"`
		ModelID   string `json:"modelId"`
		ModelFile string `json:"modelFile"`
		SpeakerID string `json:"speakerId"`
	} `json:"voice"`
	Chat struct {
		Type         string `json:"type"`
		Model        string `json:"model"`
		SystemPrompt string `json:"systemPrompt"`
	} `json:"chat"`
}

type ChatMessage struct {
	Chat []openai.ChatCompletionMessage
}
