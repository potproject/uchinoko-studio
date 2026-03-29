package db

import (
	"github.com/potproject/uchinoko-studio/data"
	"github.com/potproject/uchinoko-studio/envgen"
)

type envConfigRow struct {
	ID                       int    `db:"id"`
	OpenAISpeechToTextAPIKey string `db:"openai_speech_to_text_api_key"`
	GoogleSpeechToTextAPIKey string `db:"google_speech_to_text_api_key"`
	VoskServerEndpoint       string `db:"vosk_server_endpoint"`
	OpenAIAPIKey             string `db:"openai_api_key"`
	AnthropicAPIKey          string `db:"anthropic_api_key"`
	DeepSeekAPIKey           string `db:"deepseek_api_key"`
	GeminiAPIKey             string `db:"gemini_api_key"`
	OpenAILocalAPIKey        string `db:"openai_local_api_key"`
	OpenAILocalAPIEndpoint   string `db:"openai_local_api_endpoint"`
	VoicevoxEndpoint         string `db:"voicevox_endpoint"`
	BertVITS2Endpoint        string `db:"bertvits2_endpoint"`
	IrodoriTTSEndpoint       string `db:"irodori_tts_endpoint"`
	NijiVoiceAPIKey          string `db:"nijivoice_api_key"`
	StyleBertVIT2Endpoint    string `db:"stylebertvit2_endpoint"`
	GoogleTextToSpeechAPIKey string `db:"google_text_to_speech_api_key"`
	OpenAISpeechAPIKey       string `db:"openai_speech_api_key"`
}

func (r envConfigRow) toConfig() data.EnvConfig {
	return data.EnvConfig{
		OPENAI_SPEECH_TO_TEXT_API_KEY: r.OpenAISpeechToTextAPIKey,
		GOOGLE_SPEECH_TO_TEXT_API_KEY: r.GoogleSpeechToTextAPIKey,
		VOSK_SERVER_ENDPOINT:          r.VoskServerEndpoint,
		OPENAI_API_KEY:                r.OpenAIAPIKey,
		ANTHROPIC_API_KEY:             r.AnthropicAPIKey,
		DEEPSEEK_API_KEY:              r.DeepSeekAPIKey,
		GEMINI_API_KEY:                r.GeminiAPIKey,
		OPENAI_LOCAL_API_KEY:          r.OpenAILocalAPIKey,
		OPENAI_LOCAL_API_ENDPOINT:     r.OpenAILocalAPIEndpoint,
		VOICEVOX_ENDPOINT:             r.VoicevoxEndpoint,
		BERTVITS2_ENDPOINT:            r.BertVITS2Endpoint,
		IRODORI_TTS_ENDPOINT:          r.IrodoriTTSEndpoint,
		NIJIVOICE_API_KEY:             r.NijiVoiceAPIKey,
		STYLEBERTVIT2_ENDPOINT:        r.StyleBertVIT2Endpoint,
		GOOGLE_TEXT_TO_SPEECH_API_KEY: r.GoogleTextToSpeechAPIKey,
		OPENAI_SPEECH_API_KEY:         r.OpenAISpeechAPIKey,
	}
}

func newEnvConfigRow(config data.EnvConfig) envConfigRow {
	return envConfigRow{
		ID:                       1,
		OpenAISpeechToTextAPIKey: config.OPENAI_SPEECH_TO_TEXT_API_KEY,
		GoogleSpeechToTextAPIKey: config.GOOGLE_SPEECH_TO_TEXT_API_KEY,
		VoskServerEndpoint:       config.VOSK_SERVER_ENDPOINT,
		OpenAIAPIKey:             config.OPENAI_API_KEY,
		AnthropicAPIKey:          config.ANTHROPIC_API_KEY,
		DeepSeekAPIKey:           config.DEEPSEEK_API_KEY,
		GeminiAPIKey:             config.GEMINI_API_KEY,
		OpenAILocalAPIKey:        config.OPENAI_LOCAL_API_KEY,
		OpenAILocalAPIEndpoint:   config.OPENAI_LOCAL_API_ENDPOINT,
		VoicevoxEndpoint:         config.VOICEVOX_ENDPOINT,
		BertVITS2Endpoint:        config.BERTVITS2_ENDPOINT,
		IrodoriTTSEndpoint:       config.IRODORI_TTS_ENDPOINT,
		NijiVoiceAPIKey:          config.NIJIVOICE_API_KEY,
		StyleBertVIT2Endpoint:    config.STYLEBERTVIT2_ENDPOINT,
		GoogleTextToSpeechAPIKey: config.GOOGLE_TEXT_TO_SPEECH_API_KEY,
		OpenAISpeechAPIKey:       config.OPENAI_SPEECH_API_KEY,
	}
}

func envInitConfig() data.EnvConfig {
	return data.EnvConfig{
		OPENAI_SPEECH_TO_TEXT_API_KEY: "",
		GOOGLE_SPEECH_TO_TEXT_API_KEY: "",
		VOSK_SERVER_ENDPOINT:          "",
		OPENAI_API_KEY:                "",
		ANTHROPIC_API_KEY:             "",
		DEEPSEEK_API_KEY:              "",
		GEMINI_API_KEY:                "",
		OPENAI_LOCAL_API_KEY:          "",
		OPENAI_LOCAL_API_ENDPOINT:     "",
		VOICEVOX_ENDPOINT:             "",
		BERTVITS2_ENDPOINT:            "",
		IRODORI_TTS_ENDPOINT:          "",
		NIJIVOICE_API_KEY:             "",
		STYLEBERTVIT2_ENDPOINT:        "",
		GOOGLE_TEXT_TO_SPEECH_API_KEY: "",
		OPENAI_SPEECH_API_KEY:         "",
	}
}

func GetEnvConfig() (data.EnvConfig, error) {
	var row envConfigRow
	err := db.Get(&row, "SELECT * FROM env_config WHERE id = 1")
	if isNotFound(err) {
		return envInitConfig(), nil
	}
	if err != nil {
		return data.EnvConfig{}, err
	}

	return row.toConfig(), nil
}

func PutEnvConfig(config data.EnvConfig) error {
	row := newEnvConfigRow(config)

	_, err := db.NamedExec(`
		INSERT INTO env_config (
			id,
			openai_speech_to_text_api_key,
			google_speech_to_text_api_key,
			vosk_server_endpoint,
			openai_api_key,
			anthropic_api_key,
			deepseek_api_key,
			gemini_api_key,
			openai_local_api_key,
			openai_local_api_endpoint,
			voicevox_endpoint,
			bertvits2_endpoint,
			irodori_tts_endpoint,
			nijivoice_api_key,
			stylebertvit2_endpoint,
			google_text_to_speech_api_key,
			openai_speech_api_key
		) VALUES (
			:id,
			:openai_speech_to_text_api_key,
			:google_speech_to_text_api_key,
			:vosk_server_endpoint,
			:openai_api_key,
			:anthropic_api_key,
			:deepseek_api_key,
			:gemini_api_key,
			:openai_local_api_key,
			:openai_local_api_endpoint,
			:voicevox_endpoint,
			:bertvits2_endpoint,
			:irodori_tts_endpoint,
			:nijivoice_api_key,
			:stylebertvit2_endpoint,
			:google_text_to_speech_api_key,
			:openai_speech_api_key
		)
		ON CONFLICT(id) DO UPDATE SET
			openai_speech_to_text_api_key = excluded.openai_speech_to_text_api_key,
			google_speech_to_text_api_key = excluded.google_speech_to_text_api_key,
			vosk_server_endpoint = excluded.vosk_server_endpoint,
			openai_api_key = excluded.openai_api_key,
			anthropic_api_key = excluded.anthropic_api_key,
			deepseek_api_key = excluded.deepseek_api_key,
			gemini_api_key = excluded.gemini_api_key,
			openai_local_api_key = excluded.openai_local_api_key,
			openai_local_api_endpoint = excluded.openai_local_api_endpoint,
			voicevox_endpoint = excluded.voicevox_endpoint,
			bertvits2_endpoint = excluded.bertvits2_endpoint,
			irodori_tts_endpoint = excluded.irodori_tts_endpoint,
			nijivoice_api_key = excluded.nijivoice_api_key,
			stylebertvit2_endpoint = excluded.stylebertvit2_endpoint,
			google_text_to_speech_api_key = excluded.google_text_to_speech_api_key,
			openai_speech_api_key = excluded.openai_speech_api_key
	`, row)
	return err
}

func LoadEnvConfig() error {
	config, err := GetEnvConfig()
	if err != nil {
		return err
	}

	if config.OPENAI_SPEECH_TO_TEXT_API_KEY != "" {
		envgen.Set().OPENAI_SPEECH_API_KEY(config.OPENAI_SPEECH_TO_TEXT_API_KEY)
	}
	if config.GOOGLE_SPEECH_TO_TEXT_API_KEY != "" {
		envgen.Set().GOOGLE_SPEECH_TO_TEXT_API_KEY(config.GOOGLE_SPEECH_TO_TEXT_API_KEY)
	}
	if config.VOSK_SERVER_ENDPOINT != "" {
		envgen.Set().VOSK_SERVER_ENDPOINT(config.VOSK_SERVER_ENDPOINT)
	}
	if config.OPENAI_API_KEY != "" {
		envgen.Set().OPENAI_API_KEY(config.OPENAI_API_KEY)
	}
	if config.ANTHROPIC_API_KEY != "" {
		envgen.Set().ANTHROPIC_API_KEY(config.ANTHROPIC_API_KEY)
	}
	if config.DEEPSEEK_API_KEY != "" {
		envgen.Set().DEEPSEEK_API_KEY(config.DEEPSEEK_API_KEY)
	}
	if config.GEMINI_API_KEY != "" {
		envgen.Set().GEMINI_API_KEY(config.GEMINI_API_KEY)
	}
	if config.OPENAI_LOCAL_API_KEY != "" {
		envgen.Set().OPENAI_LOCAL_API_KEY(config.OPENAI_LOCAL_API_KEY)
	}
	if config.OPENAI_LOCAL_API_ENDPOINT != "" {
		envgen.Set().OPENAI_LOCAL_API_ENDPOINT(config.OPENAI_LOCAL_API_ENDPOINT)
	}
	if config.VOICEVOX_ENDPOINT != "" {
		envgen.Set().VOICEVOX_ENDPOINT(config.VOICEVOX_ENDPOINT)
	}
	if config.BERTVITS2_ENDPOINT != "" {
		envgen.Set().BERTVITS2_ENDPOINT(config.BERTVITS2_ENDPOINT)
	}
	if config.IRODORI_TTS_ENDPOINT != "" {
		envgen.Set().IRODORI_TTS_ENDPOINT(config.IRODORI_TTS_ENDPOINT)
	}
	if config.STYLEBERTVIT2_ENDPOINT != "" {
		envgen.Set().STYLEBERTVIT2_ENDPOINT(config.STYLEBERTVIT2_ENDPOINT)
	}
	if config.GOOGLE_TEXT_TO_SPEECH_API_KEY != "" {
		envgen.Set().GOOGLE_TEXT_TO_SPEECH_API_KEY(config.GOOGLE_TEXT_TO_SPEECH_API_KEY)
	}
	if config.OPENAI_SPEECH_API_KEY != "" {
		envgen.Set().OPENAI_SPEECH_API_KEY(config.OPENAI_SPEECH_API_KEY)
	}

	return nil
}
