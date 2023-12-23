package db

import (
	"encoding/json"

	"github.com/potproject/uchinoko-studio/data"
	"github.com/syndtr/goleveldb/leveldb"
)

func initConfig() data.Config {
	return data.Config{}
}

func GetConfig(id string) (data.Config, error) {
	key := []byte(id + "/config")
	value, err := get(key)
	if err == leveldb.ErrNotFound {
		return initConfig(), nil
	} else if err != nil {
		return data.Config{}, err
	}
	var config data.Config
	err = json.Unmarshal(value, &config)
	if err != nil {
		return data.Config{}, err
	}
	return config, nil
}

func PutConfig(id string, config data.Config) error {
	key := []byte(id + "/config")
	value, err := json.Marshal(config)
	if err != nil {
		return err
	}
	return put(key, value)
}
