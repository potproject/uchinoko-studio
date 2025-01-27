package data

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
	ID   string `json:"id"`
	Name string `json:"name"`
}

type CharacterConfigVoice struct {
	Name                string                         `json:"name"`
	Type                string                         `json:"type"`
	Identification      string                         `json:"identification"`
	ModelID             string                         `json:"modelId"`
	ModelFile           string                         `json:"modelFile"`
	SpeakerID           string                         `json:"speakerId"`
	Image               string                         `json:"image"`
	BackgroundImagePath string                         `json:"backgroundImagePath"`
	Behavior            []CharacterConfigVoiceBehavior `json:"behavior"`
}

type CharacterConfigVoiceBehavior struct {
	Identification string `json:"identification"`
	ImagePath      string `json:"imagePath"`
}

type CharacterConfigChat struct {
	Type         string                   `json:"type"`
	Model        string                   `json:"model"`
	SystemPrompt string                   `json:"systemPrompt"`
	Temperature  TemperatureConfig        `json:"temperature"`
	MaxHistory   int64                    `json:"maxHistory"`
	Limit        CharacterConfigChatLimit `json:"limit"`
}

type TemperatureConfig struct {
	Enable bool    `json:"enable"`
	Value  float32 `json:"value"`
}

type CharacterConfigChatLimit struct {
	Day    CharacterConfigChatLimitType `json:"day"`
	Hour   CharacterConfigChatLimitType `json:"hour"`
	Minute CharacterConfigChatLimitType `json:"minute"`
}

type CharacterConfigChatLimitType struct {
	Request int64 `json:"request"`
	Token   int64 `json:"token"`
}
