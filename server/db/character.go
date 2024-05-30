package db

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/potproject/uchinoko-studio/data"
	"github.com/syndtr/goleveldb/leveldb"
)

func CharacterInitConfig() data.CharacterConfig {
	return data.CharacterConfig{
		General: data.CharacterConfigGeneral{
			ID:   uuid.New().String(),
			Name: "Default",
		},
		MultiVoice: false,
		Voice: []data.CharacterConfigVoice{
			{
				Type:                "voicevox",
				Identification:      "",
				ModelID:             "",
				ModelFile:           "",
				SpeakerID:           "1",
				Image:               "default.png",
				BackgroundImagePath: "",
				Behavior:            []data.CharacterConfigVoiceBehavior{},
			},
		},
		Chat: data.CharacterConfigChat{
			Type:         "openai",
			Model:        "gpt-3.5-turbo",
			SystemPrompt: "The following is a conversation with an AI assistant. The assistant is helpful, creative, clever, and very friendly.",
			MaxHistory:   0,
			Limit: data.CharacterConfigChatLimit{
				Day:    data.CharacterConfigChatLimitType{Request: 0, Token: 0},
				Hour:   data.CharacterConfigChatLimitType{Request: 0, Token: 0},
				Minute: data.CharacterConfigChatLimitType{Request: 0, Token: 0},
			},
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
