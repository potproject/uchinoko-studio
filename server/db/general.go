package db

import (
	"encoding/json"

	"github.com/potproject/uchinoko-studio/data"
	"github.com/syndtr/goleveldb/leveldb"
)

func generalInitConfig() data.GeneralConfig {
	return data.GeneralConfig{
		Background:            "blue",
		Language:              "ja-JP",
		SoundEffect:           true,
		CharacterOutputChange: false,
		Transcription: struct {
			Type        string `json:"type"`
			Method      string `json:"method"`
			AutoSetting struct {
				Threshold       float64 `json:"threshold"`
				SilentThreshold float64 `json:"silentThreshold"`
				AudioMinLength  float64 `json:"audioMinLength"`
			} `json:"autoSetting"`
		}{
			Type:   "openai_speech_to_text",
			Method: "auto",
			AutoSetting: struct {
				Threshold       float64 `json:"threshold"`
				SilentThreshold float64 `json:"silentThreshold"`
				AudioMinLength  float64 `json:"audioMinLength"`
			}{
				Threshold:       0.02,
				SilentThreshold: 1,
				AudioMinLength:  1.3,
			},
		},
	}
}

const generalConfigPrefix = "general_config"

func GetGeneralConfig() (data.GeneralConfig, error) {
	key := []byte(generalConfigPrefix)
	value, err := get(key)
	if err == leveldb.ErrNotFound {
		return generalInitConfig(), nil
	} else if err != nil {
		return data.GeneralConfig{}, err
	}
	var config data.GeneralConfig
	err = json.Unmarshal(value, &config)
	if err != nil {
		return data.GeneralConfig{}, err
	}
	return config, nil
}

func PutGeneralConfig(config data.GeneralConfig) error {
	key := []byte(generalConfigPrefix)
	value, err := json.Marshal(config)
	if err != nil {
		return err
	}
	return put(key, value)
}
