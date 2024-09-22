package db

import (
	"encoding/json"

	"github.com/potproject/uchinoko-studio/data"
	"github.com/potproject/uchinoko-studio/envgen"
	"github.com/syndtr/goleveldb/leveldb"
)

// 環境変数の初期設定（必要に応じてデフォルト値を設定）
func envInitConfig() data.EnvConfig {
	return data.EnvConfig{
		OPENAI_SPEECH_TO_TEXT_API_KEY: "",
		GOOGLE_SPEECH_TO_TEXT_API_KEY: "",
		VOSK_SERVER_ENDPOINT:          "",
		OPENAI_API_KEY:                "",
		ANTHROPIC_API_KEY:             "",
		COHERE_API_KEY:                "",
		GEMINI_API_KEY:                "",
		OPENAI_LOCAL_API_KEY:          "",
		OPENAI_LOCAL_API_ENDPOINT:     "",
		VOICEVOX_ENDPOINT:             "",
		BERTVITS2_ENDPOINT:            "",
		STYLEBERTVIT2_ENDPOINT:        "",
		GOOGLE_TEXT_TO_SPEECH_API_KEY: "",
		OPENAI_SPEECH_API_KEY:         "",
	}
}

const envConfigPrefix = "env_config"

func GetEnvConfig() (data.EnvConfig, error) {
	key := []byte(envConfigPrefix)
	value, err := get(key)
	if err == leveldb.ErrNotFound {
		return envInitConfig(), nil
	} else if err != nil {
		return data.EnvConfig{}, err
	}
	var config data.EnvConfig
	err = json.Unmarshal(value, &config)
	if err != nil {
		return data.EnvConfig{}, err
	}
	return config, nil
}

func PutEnvConfig(config data.EnvConfig) error {
	key := []byte(envConfigPrefix)
	value, err := json.Marshal(config)
	if err != nil {
		return err
	}
	return put(key, value)
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
	if config.COHERE_API_KEY != "" {
		envgen.Set().COHERE_API_KEY(config.COHERE_API_KEY)
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
