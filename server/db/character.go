package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/potproject/uchinoko-studio/data"
	"github.com/potproject/uchinoko-studio/db/sqlcgen"
)

func characterConfigFromRow(row sqlcgen.Character) (data.CharacterConfig, error) {
	voice, err := unmarshalJSONString[[]data.CharacterConfigVoice](row.VoiceJson)
	if err != nil {
		return data.CharacterConfig{}, err
	}

	chat, err := unmarshalJSONString[data.CharacterConfigChat](row.ChatJson)
	if err != nil {
		return data.CharacterConfig{}, err
	}

	return data.CharacterConfig{
		General: data.CharacterConfigGeneral{
			ID:   row.ID,
			Name: row.Name,
		},
		MultiVoice: intToBool(row.MultiVoice),
		Voice:      voice,
		Chat:       chat,
	}, nil
}

func newCharacterParams(config data.CharacterConfig) (sqlcgen.UpsertCharacterParams, error) {
	voiceJSON, err := marshalJSONString(config.Voice)
	if err != nil {
		return sqlcgen.UpsertCharacterParams{}, err
	}

	chatJSON, err := marshalJSONString(config.Chat)
	if err != nil {
		return sqlcgen.UpsertCharacterParams{}, err
	}

	return sqlcgen.UpsertCharacterParams{
		ID:         config.General.ID,
		Name:       config.General.Name,
		MultiVoice: boolToInt(config.MultiVoice),
		VoiceJson:  voiceJSON,
		ChatJson:   chatJSON,
	}, nil
}

func CharacterInitConfig() data.CharacterConfig {
	return data.CharacterConfig{
		General: data.CharacterConfigGeneral{
			ID:   uuid.New().String(),
			Name: "Default",
		},
		MultiVoice: false,
		Voice: []data.CharacterConfigVoice{
			{
				Name:                "Default",
				Type:                "voicevox",
				Identification:      "",
				ModelID:             "",
				ModelFile:           "",
				SpeakerID:           "1",
				ReferenceAudioPath:  "",
				Image:               "default.png",
				BackgroundImagePath: "",
				Behavior:            []data.CharacterConfigVoiceBehavior{},
			},
		},
		Chat: data.CharacterConfigChat{
			Type:         "openai",
			Model:        "gpt-4o-mini",
			SystemPrompt: "The following is a conversation with an AI assistant. The assistant is helpful, creative, clever, and very friendly.",
			Temperature: data.TemperatureConfig{
				Enable: false,
				Value:  0.0,
			},
			MaxHistory: 0,
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

func GetCharacterConfigList() (data.CharacterConfigList, error) {
	rows, err := queries.ListCharacters(context.Background())
	if err != nil {
		return data.CharacterConfigList{}, err
	}

	configs := make([]data.CharacterConfig, 0, len(rows))
	for _, row := range rows {
		config, err := characterConfigFromRow(row)
		if err != nil {
			return data.CharacterConfigList{}, err
		}
		configs = append(configs, config)
	}

	return data.CharacterConfigList{Characters: configs}, nil
}

func GetCharacterConfig(id string) (data.CharacterConfig, error) {
	row, err := queries.GetCharacter(context.Background(), id)
	if isNotFound(err) {
		return data.CharacterConfig{}, nil
	}
	if err != nil {
		return data.CharacterConfig{}, err
	}

	return characterConfigFromRow(row)
}

func DeleteCharacterConfig(id string) error {
	return queries.DeleteCharacter(context.Background(), id)
}

func PutCharacterConfig(id string, config data.CharacterConfig) error {
	config.General.ID = id

	row, err := newCharacterParams(config)
	if err != nil {
		return err
	}

	return queries.UpsertCharacter(context.Background(), row)
}
