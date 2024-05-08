package db

import (
	"encoding/json"

	"github.com/potproject/uchinoko-studio/data"
	"github.com/syndtr/goleveldb/leveldb"
)

func generalInitConfig() data.GeneralConfig {
	return data.GeneralConfig{
		Transcription: struct {
			Type   string `json:"type"`
			Method string `json:"method"`
		}{
			Type:   "whisper",
			Method: "auto",
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
