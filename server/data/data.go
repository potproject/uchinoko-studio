package data

type GeneralConfig struct {
	Transcription struct {
		Type        string `json:"type"`
		Method      string `json:"method"`
		AutoSetting struct {
			Threshold       float64 `json:"threshold"`
			SilentThreshold float64 `json:"silentThreshold"`
			AudioMinLength  float64 `json:"audioMinLength"`
		} `json:"autoSetting"`
	} `json:"transcription"`
}

type CharacterConfigList struct {
	Characters []CharacterConfig `json:"characters"`
}

type CharacterConfig struct {
	General    CharacterConfigGeneral `json:"general"`
	MultiVoice bool                   `json:"multiVoice"`
	Voice      []CharacterConfigVoice `json:"voice"`
	Chat       CharacterConfigChat    `json:"chat"`
}

type CharacterConfigGeneral struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Image string `json:"image"`
}

type CharacterConfigVoice struct {
	Type           string `json:"type"`
	Identification string `json:"identification"`
	ModelID        string `json:"modelId"`
	ModelFile      string `json:"modelFile"`
	SpeakerID      string `json:"speakerId"`
	Image          string `json:"image"`
}

type CharacterConfigChat struct {
	Type         string `json:"type"`
	Model        string `json:"model"`
	SystemPrompt string `json:"systemPrompt"`
}

type ChatMessage struct {
	Chat []ChatCompletionMessage
}

type ChatCompletionMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
