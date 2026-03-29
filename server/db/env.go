package db

import (
	"context"

	"github.com/potproject/uchinoko-studio/data"
	"github.com/potproject/uchinoko-studio/db/sqlcgen"
	"github.com/potproject/uchinoko-studio/envgen"
)

func envConfigFromRow(row sqlcgen.EnvConfig) data.EnvConfig {
	return data.EnvConfig{
		OPENAI_SPEECH_TO_TEXT_API_KEY: row.OpenaiSpeechToTextApiKey,
		GOOGLE_SPEECH_TO_TEXT_API_KEY: row.GoogleSpeechToTextApiKey,
		VOSK_SERVER_ENDPOINT:          row.VoskServerEndpoint,
		OPENAI_API_KEY:                row.OpenaiApiKey,
		ANTHROPIC_API_KEY:             row.AnthropicApiKey,
		DEEPSEEK_API_KEY:              row.DeepseekApiKey,
		GEMINI_API_KEY:                row.GeminiApiKey,
		OPENAI_LOCAL_API_KEY:          row.OpenaiLocalApiKey,
		OPENAI_LOCAL_API_ENDPOINT:     row.OpenaiLocalApiEndpoint,
		VOICEVOX_ENDPOINT:             row.VoicevoxEndpoint,
		BERTVITS2_ENDPOINT:            row.Bertvits2Endpoint,
		IRODORI_TTS_ENDPOINT:          row.IrodoriTtsEndpoint,
		NIJIVOICE_API_KEY:             row.NijivoiceApiKey,
		STYLEBERTVIT2_ENDPOINT:        row.Stylebertvit2Endpoint,
		GOOGLE_TEXT_TO_SPEECH_API_KEY: row.GoogleTextToSpeechApiKey,
		OPENAI_SPEECH_API_KEY:         row.OpenaiSpeechApiKey,
	}
}

func newEnvConfigParams(config data.EnvConfig) sqlcgen.UpsertEnvConfigParams {
	return sqlcgen.UpsertEnvConfigParams{
		ID:                       1,
		OpenaiSpeechToTextApiKey: config.OPENAI_SPEECH_TO_TEXT_API_KEY,
		GoogleSpeechToTextApiKey: config.GOOGLE_SPEECH_TO_TEXT_API_KEY,
		VoskServerEndpoint:       config.VOSK_SERVER_ENDPOINT,
		OpenaiApiKey:             config.OPENAI_API_KEY,
		AnthropicApiKey:          config.ANTHROPIC_API_KEY,
		DeepseekApiKey:           config.DEEPSEEK_API_KEY,
		GeminiApiKey:             config.GEMINI_API_KEY,
		OpenaiLocalApiKey:        config.OPENAI_LOCAL_API_KEY,
		OpenaiLocalApiEndpoint:   config.OPENAI_LOCAL_API_ENDPOINT,
		VoicevoxEndpoint:         config.VOICEVOX_ENDPOINT,
		Bertvits2Endpoint:        config.BERTVITS2_ENDPOINT,
		IrodoriTtsEndpoint:       config.IRODORI_TTS_ENDPOINT,
		NijivoiceApiKey:          config.NIJIVOICE_API_KEY,
		Stylebertvit2Endpoint:    config.STYLEBERTVIT2_ENDPOINT,
		GoogleTextToSpeechApiKey: config.GOOGLE_TEXT_TO_SPEECH_API_KEY,
		OpenaiSpeechApiKey:       config.OPENAI_SPEECH_API_KEY,
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
	row, err := queries.GetEnvConfig(context.Background())
	if isNotFound(err) {
		return envInitConfig(), nil
	}
	if err != nil {
		return data.EnvConfig{}, err
	}

	return envConfigFromRow(row), nil
}

func PutEnvConfig(config data.EnvConfig) error {
	return queries.UpsertEnvConfig(context.Background(), newEnvConfigParams(config))
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
