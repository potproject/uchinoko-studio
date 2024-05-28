package data

type GeneralConfig struct {
	Language      string `json:"language"`
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
