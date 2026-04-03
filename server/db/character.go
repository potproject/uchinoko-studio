package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/potproject/uchinoko-studio/data"
	"github.com/potproject/uchinoko-studio/db/sqlcgen"
)

func characterConfigFromRows(row sqlcgen.Character, chatRow sqlcgen.CharacterChatSetting, limitRows []sqlcgen.CharacterChatLimit, voiceRows []sqlcgen.CharacterVoice, behaviorRows []sqlcgen.CharacterVoiceBehavior) (data.CharacterConfig, error) {
	voiceByIndex := make(map[int64]*data.CharacterConfigVoice, len(voiceRows))
	voices := make([]data.CharacterConfigVoice, 0, len(voiceRows))
	for _, voiceRow := range voiceRows {
		voice := data.CharacterConfigVoice{
			Name:                voiceRow.Name,
			Type:                voiceRow.Type,
			Identification:      voiceRow.Identification,
			ModelID:             voiceRow.ModelID,
			ModelFile:           voiceRow.ModelFile,
			SpeakerID:           voiceRow.SpeakerID,
			ReferenceAudioPath:  voiceRow.ReferenceAudioPath,
			Image:               voiceRow.Image,
			BackgroundImagePath: voiceRow.BackgroundImagePath,
		}
		voices = append(voices, voice)
		voiceByIndex[voiceRow.VoiceIndex] = &voices[len(voices)-1]
	}

	for _, behaviorRow := range behaviorRows {
		voice := voiceByIndex[behaviorRow.VoiceIndex]
		if voice == nil {
			return data.CharacterConfig{}, fmt.Errorf("character %s has behavior for missing voice index %d", row.ID, behaviorRow.VoiceIndex)
		}
		voice.Behavior = append(voice.Behavior, data.CharacterConfigVoiceBehavior{
			Identification: behaviorRow.Identification,
			ImagePath:      behaviorRow.ImagePath,
		})
	}

	limit := data.CharacterConfigChatLimit{}
	for _, limitRow := range limitRows {
		limitType := data.CharacterConfigChatLimitType{
			Request: limitRow.RequestLimit,
			Token:   limitRow.TokenLimit,
		}
		switch limitRow.Window {
		case "day":
			limit.Day = limitType
		case "hour":
			limit.Hour = limitType
		case "minute":
			limit.Minute = limitType
		default:
			return data.CharacterConfig{}, fmt.Errorf("character %s has unsupported chat limit window %q", row.ID, limitRow.Window)
		}
	}

	return data.CharacterConfig{
		General: data.CharacterConfigGeneral{
			ID:   row.ID,
			Name: row.Name,
		},
		MultiVoice: intToBool(row.MultiVoice),
		Voice:      voices,
		Chat: data.CharacterConfigChat{
			Type:         chatRow.Type,
			Model:        chatRow.Model,
			SystemPrompt: chatRow.SystemPrompt,
			Temperature: data.TemperatureConfig{
				Enable: intToBool(chatRow.TemperatureEnable),
				Value:  float32(chatRow.TemperatureValue),
			},
			MaxHistory: chatRow.MaxHistory,
			Limit:      limit,
		},
		Memory: memoryConfigInit(),
	}, nil
}

func loadCharacterConfig(ctx context.Context, q *sqlcgen.Queries, row sqlcgen.Character) (data.CharacterConfig, error) {
	chatRow, err := q.GetCharacterChatSetting(ctx, row.ID)
	if err != nil {
		return data.CharacterConfig{}, err
	}

	limitRows, err := q.ListCharacterChatLimits(ctx, row.ID)
	if err != nil {
		return data.CharacterConfig{}, err
	}

	voiceRows, err := q.ListCharacterVoices(ctx, row.ID)
	if err != nil {
		return data.CharacterConfig{}, err
	}

	behaviorRows, err := q.ListCharacterVoiceBehaviors(ctx, row.ID)
	if err != nil {
		return data.CharacterConfig{}, err
	}

	return characterConfigFromRows(row, chatRow, limitRows, voiceRows, behaviorRows)
}

func newCharacterParams(config data.CharacterConfig) sqlcgen.UpsertCharacterParams {
	return sqlcgen.UpsertCharacterParams{
		ID:         config.General.ID,
		Name:       config.General.Name,
		MultiVoice: boolToInt(config.MultiVoice),
	}
}

func newCharacterChatSettingParams(config data.CharacterConfig) sqlcgen.UpsertCharacterChatSettingParams {
	return sqlcgen.UpsertCharacterChatSettingParams{
		CharacterID:       config.General.ID,
		Type:              config.Chat.Type,
		Model:             config.Chat.Model,
		SystemPrompt:      config.Chat.SystemPrompt,
		TemperatureEnable: boolToInt(config.Chat.Temperature.Enable),
		TemperatureValue:  float64(config.Chat.Temperature.Value),
		MaxHistory:        config.Chat.MaxHistory,
	}
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
		Memory: memoryConfigInit(),
	}
}

func CharactersInitConfig() data.CharacterConfigList {
	return data.CharacterConfigList{
		Characters: []data.CharacterConfig{},
	}
}

func GetCharacterConfigList() (data.CharacterConfigList, error) {
	ctx := context.Background()

	rows, err := queries.ListCharacters(ctx)
	if err != nil {
		return data.CharacterConfigList{}, err
	}

	configs := make([]data.CharacterConfig, 0, len(rows))
	for _, row := range rows {
		config, err := loadCharacterConfig(ctx, queries, row)
		if err != nil {
			return data.CharacterConfigList{}, err
		}
		config.Memory = getCharacterMemoryConfig(row.ID)
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
	config, err := loadCharacterConfig(context.Background(), queries, row)
	if err != nil {
		return data.CharacterConfig{}, err
	}
	config.Memory = getCharacterMemoryConfig(row.ID)
	return config, nil
}

func DeleteCharacterConfig(id string) error {
	return queries.DeleteCharacter(context.Background(), id)
}

func PutCharacterConfig(id string, config data.CharacterConfig) error {
	ctx := context.Background()
	config.General.ID = id

	return withTxExec(ctx, func(tx *sql.Tx, qtx *sqlcgen.Queries) error {
		if err := qtx.UpsertCharacter(ctx, newCharacterParams(config)); err != nil {
			return err
		}
		if err := qtx.UpsertCharacterChatSetting(ctx, newCharacterChatSettingParams(config)); err != nil {
			return err
		}
		if err := qtx.DeleteCharacterChatLimits(ctx, config.General.ID); err != nil {
			return err
		}

		limits := []sqlcgen.InsertCharacterChatLimitParams{
			{
				CharacterID:  config.General.ID,
				Window:       "day",
				RequestLimit: config.Chat.Limit.Day.Request,
				TokenLimit:   config.Chat.Limit.Day.Token,
			},
			{
				CharacterID:  config.General.ID,
				Window:       "hour",
				RequestLimit: config.Chat.Limit.Hour.Request,
				TokenLimit:   config.Chat.Limit.Hour.Token,
			},
			{
				CharacterID:  config.General.ID,
				Window:       "minute",
				RequestLimit: config.Chat.Limit.Minute.Request,
				TokenLimit:   config.Chat.Limit.Minute.Token,
			},
		}
		for _, limit := range limits {
			if err := qtx.InsertCharacterChatLimit(ctx, limit); err != nil {
				return err
			}
		}

		if err := qtx.DeleteCharacterVoices(ctx, config.General.ID); err != nil {
			return err
		}
		for voiceIndex, voice := range config.Voice {
			if err := qtx.InsertCharacterVoice(ctx, sqlcgen.InsertCharacterVoiceParams{
				CharacterID:         config.General.ID,
				VoiceIndex:          int64(voiceIndex),
				Name:                voice.Name,
				Type:                voice.Type,
				Identification:      voice.Identification,
				ModelID:             voice.ModelID,
				ModelFile:           voice.ModelFile,
				SpeakerID:           voice.SpeakerID,
				ReferenceAudioPath:  voice.ReferenceAudioPath,
				Image:               voice.Image,
				BackgroundImagePath: voice.BackgroundImagePath,
			}); err != nil {
				return err
			}

			for behaviorIndex, behavior := range voice.Behavior {
				if err := qtx.InsertCharacterVoiceBehavior(ctx, sqlcgen.InsertCharacterVoiceBehaviorParams{
					CharacterID:    config.General.ID,
					VoiceIndex:     int64(voiceIndex),
					BehaviorIndex:  int64(behaviorIndex),
					Identification: behavior.Identification,
					ImagePath:      behavior.ImagePath,
				}); err != nil {
					return err
				}
			}
		}

		if err := putCharacterMemoryConfigTx(ctx, tx, config.General.ID, config.Memory); err != nil {
			return err
		}

		return nil
	})
}
