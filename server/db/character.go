package db

import (
	"github.com/google/uuid"
	"github.com/potproject/uchinoko-studio/data"
)

type characterRow struct {
	ID         string `db:"id"`
	Name       string `db:"name"`
	MultiVoice int    `db:"multi_voice"`
	VoiceJSON  string `db:"voice_json"`
	ChatJSON   string `db:"chat_json"`
}

func (r characterRow) toConfig() (data.CharacterConfig, error) {
	voice, err := unmarshalJSONString[[]data.CharacterConfigVoice](r.VoiceJSON)
	if err != nil {
		return data.CharacterConfig{}, err
	}

	chat, err := unmarshalJSONString[data.CharacterConfigChat](r.ChatJSON)
	if err != nil {
		return data.CharacterConfig{}, err
	}

	return data.CharacterConfig{
		General: data.CharacterConfigGeneral{
			ID:   r.ID,
			Name: r.Name,
		},
		MultiVoice: intToBool(r.MultiVoice),
		Voice:      voice,
		Chat:       chat,
	}, nil
}

func newCharacterRow(config data.CharacterConfig) (characterRow, error) {
	voiceJSON, err := marshalJSONString(config.Voice)
	if err != nil {
		return characterRow{}, err
	}

	chatJSON, err := marshalJSONString(config.Chat)
	if err != nil {
		return characterRow{}, err
	}

	return characterRow{
		ID:         config.General.ID,
		Name:       config.General.Name,
		MultiVoice: boolToInt(config.MultiVoice),
		VoiceJSON:  voiceJSON,
		ChatJSON:   chatJSON,
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
	var rows []characterRow
	if err := db.Select(&rows, "SELECT * FROM characters ORDER BY name, id"); err != nil {
		return data.CharacterConfigList{}, err
	}

	configs := make([]data.CharacterConfig, 0, len(rows))
	for _, row := range rows {
		config, err := row.toConfig()
		if err != nil {
			return data.CharacterConfigList{}, err
		}
		configs = append(configs, config)
	}

	return data.CharacterConfigList{Characters: configs}, nil
}

func GetCharacterConfig(id string) (data.CharacterConfig, error) {
	var row characterRow
	err := db.Get(&row, "SELECT * FROM characters WHERE id = ?", id)
	if isNotFound(err) {
		return data.CharacterConfig{}, nil
	}
	if err != nil {
		return data.CharacterConfig{}, err
	}

	return row.toConfig()
}

func DeleteCharacterConfig(id string) error {
	_, err := db.Exec("DELETE FROM characters WHERE id = ?", id)
	return err
}

func PutCharacterConfig(id string, config data.CharacterConfig) error {
	config.General.ID = id

	row, err := newCharacterRow(config)
	if err != nil {
		return err
	}

	_, err = db.NamedExec(`
		INSERT INTO characters (id, name, multi_voice, voice_json, chat_json)
		VALUES (:id, :name, :multi_voice, :voice_json, :chat_json)
		ON CONFLICT(id) DO UPDATE SET
			name = excluded.name,
			multi_voice = excluded.multi_voice,
			voice_json = excluded.voice_json,
			chat_json = excluded.chat_json
	`, row)
	return err
}
