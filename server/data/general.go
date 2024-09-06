package data

type GeneralConfig struct {
	Language              string `json:"language"`
	SoundEffect           bool   `json:"soundEffect"`
	CharacterOutputChange bool   `json:"characterOutputChange"`
	Transcription         struct {
		Type        string `json:"type"`
		Method      string `json:"method"`
		AutoSetting struct {
			Threshold       float64 `json:"threshold"`
			SilentThreshold float64 `json:"silentThreshold"`
			AudioMinLength  float64 `json:"audioMinLength"`
		} `json:"autoSetting"`
	} `json:"transcription"`
}
