package data

type EnvConfig struct {
	OPENAI_SPEECH_TO_TEXT_API_KEY string `json:"OPENAI_SPEECH_TO_TEXT_API_KEY"`
	GOOGLE_SPEECH_TO_TEXT_API_KEY string `json:"GOOGLE_SPEECH_TO_TEXT_API_KEY"`
	VOSK_SERVER_ENDPOINT          string `json:"VOSK_SERVER_ENDPOINT"`
	OPENAI_API_KEY                string `json:"OPENAI_API_KEY"`
	ANTHROPIC_API_KEY             string `json:"ANTHROPIC_API_KEY"`
	COHERE_API_KEY                string `json:"COHERE_API_KEY"`
	GEMINI_API_KEY                string `json:"GEMINI_API_KEY"`
	OPENAI_LOCAL_API_KEY          string `json:"OPENAI_LOCAL_API_KEY"`
	OPENAI_LOCAL_API_ENDPOINT     string `json:"OPENAI_LOCAL_API_ENDPOINT"`
	VOICEVOX_ENDPOINT             string `json:"VOICEVOX_ENDPOINT"`
	BERTVITS2_ENDPOINT            string `json:"BERTVITS2_ENDPOINT"`
	NIJIVOICE_API_KEY             string `json:"NIJIVOICE_API_KEY"`
	STYLEBERTVIT2_ENDPOINT        string `json:"STYLEBERTVIT2_ENDPOINT"`
	GOOGLE_TEXT_TO_SPEECH_API_KEY string `json:"GOOGLE_TEXT_TO_SPEECH_API_KEY"`
	OPENAI_SPEECH_API_KEY         string `json:"OPENAI_SPEECH_API_KEY"`
}