package db

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/potproject/uchinoko-studio/data"
	"github.com/syndtr/goleveldb/leveldb"
)

func CharacterInitConfig() data.CharacterConfig {
	return data.CharacterConfig{
		General: struct {
			ID    string `json:"id"`
			Name  string `json:"name"`
			Image string `json:"image"`
		}{
			ID:    uuid.New().String(),
			Name:  "Default",
			Image: "default.png",
		},
		MultiVoice: false,
		Voice: []struct {
			Type           string `json:"type"`
			Identification string `json:"identification"`
			ModelID        string `json:"modelId"`
			ModelFile      string `json:"modelFile"`
			SpeakerID      string `json:"speakerId"`
		}{
			{
				Type:      "bertvits2",
				ModelID:   "0",
				ModelFile: "",
				SpeakerID: "0",
			},
		},
		Chat: struct {
			Type         string `json:"type"`
			Model        string `json:"model"`
			SystemPrompt string `json:"systemPrompt"`
		}{
			Type:         "openai",
			Model:        "gpt-3.5-turbo",
			SystemPrompt: "The following is a conversation with an AI assistant. The assistant is helpful, creative, clever, and very friendly.",
		},
	}
}

func CharactersInitConfig() data.CharacterConfigList {
	return data.CharacterConfigList{
		Characters: []data.CharacterConfig{},
	}
}

const characterConfigPrefix = "character_config"

func GetCharacterConfigList() (data.CharacterConfigList, error) {
	key := []byte(characterConfigPrefix)
	value, err := get(key)
	if err == leveldb.ErrNotFound {
		return CharactersInitConfig(), nil
	} else if err != nil {
		return data.CharacterConfigList{}, err
	}
	var config data.CharacterConfigList
	err = json.Unmarshal(value, &config)
	if err != nil {
		return data.CharacterConfigList{}, err
	}
	return config, nil
}

func GetCharacterConfig(id string) (data.CharacterConfig, error) {
	configs, err := GetCharacterConfigList()
	if err != nil {
		return data.CharacterConfig{}, err
	}
	for _, c := range configs.Characters {
		if c.General.ID == id {
			return c, nil
		}
	}
	return data.CharacterConfig{}, nil
}

func DeleteCharacterConfig(id string) error {
	configs, err := GetCharacterConfigList()
	if err != nil {
		return err
	}
	for i, c := range configs.Characters {
		if c.General.ID == id {
			configs.Characters = append(configs.Characters[:i], configs.Characters[i+1:]...)
			break
		}
	}
	key := []byte(characterConfigPrefix)
	value, err := json.Marshal(configs)
	if err != nil {
		return err
	}
	return put(key, value)
}

func PutCharacterConfig(id string, config data.CharacterConfig) error {
	configs, err := GetCharacterConfigList()
	if err != nil {
		return err
	}
	exists := false
	for i, c := range configs.Characters {
		if c.General.ID == id {
			configs.Characters[i] = config
			exists = true
			break
		}
	}
	if !exists {
		configs.Characters = append(configs.Characters, config)
	}
	key := []byte(characterConfigPrefix)
	value, err := json.Marshal(configs)
	if err != nil {
		return err
	}
	return put(key, value)
}
