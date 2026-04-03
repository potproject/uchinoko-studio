package db

import (
	"context"

	"github.com/potproject/uchinoko-studio/data"
	"github.com/potproject/uchinoko-studio/db/sqlcgen"
)

func generalConfigFromRow(row sqlcgen.GeneralConfig) data.GeneralConfig {
	return data.GeneralConfig{
		Background:            row.Background,
		Language:              row.Language,
		SoundEffect:           intToBool(row.SoundEffect),
		CharacterOutputChange: intToBool(row.CharacterOutputChange),
		EnableTTSOptimization: intToBool(row.EnableTtsOptimization),
		Transcription: struct {
			Type        string `json:"type"`
			Method      string `json:"method"`
			AutoSetting struct {
				Threshold       float64 `json:"threshold"`
				SilentThreshold float64 `json:"silentThreshold"`
				AudioMinLength  float64 `json:"audioMinLength"`
			} `json:"autoSetting"`
		}{
			Type:   row.TranscriptionType,
			Method: row.TranscriptionMethod,
			AutoSetting: struct {
				Threshold       float64 `json:"threshold"`
				SilentThreshold float64 `json:"silentThreshold"`
				AudioMinLength  float64 `json:"audioMinLength"`
			}{
				Threshold:       row.TranscriptionAutoThreshold,
				SilentThreshold: row.TranscriptionAutoSilentThreshold,
				AudioMinLength:  row.TranscriptionAutoAudioMinLength,
			},
		},
	}
}

func newGeneralConfigParams(config data.GeneralConfig) sqlcgen.UpsertGeneralConfigParams {
	return sqlcgen.UpsertGeneralConfigParams{
		ID:                               1,
		Background:                       config.Background,
		Language:                         config.Language,
		SoundEffect:                      boolToInt(config.SoundEffect),
		CharacterOutputChange:            boolToInt(config.CharacterOutputChange),
		EnableTtsOptimization:            boolToInt(config.EnableTTSOptimization),
		TranscriptionType:                config.Transcription.Type,
		TranscriptionMethod:              config.Transcription.Method,
		TranscriptionAutoThreshold:       config.Transcription.AutoSetting.Threshold,
		TranscriptionAutoSilentThreshold: config.Transcription.AutoSetting.SilentThreshold,
		TranscriptionAutoAudioMinLength:  config.Transcription.AutoSetting.AudioMinLength,
	}
}

func generalInitConfig() data.GeneralConfig {
	return data.GeneralConfig{
		Background:            "blue",
		Language:              "ja-JP",
		SoundEffect:           true,
		CharacterOutputChange: false,
		EnableTTSOptimization: false,
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

func GetGeneralConfig() (data.GeneralConfig, error) {
	row, err := queries.GetGeneralConfig(context.Background())
	if isNotFound(err) {
		return generalInitConfig(), nil
	}
	if err != nil {
		return data.GeneralConfig{}, err
	}

	return generalConfigFromRow(row), nil
}

func PutGeneralConfig(config data.GeneralConfig) error {
	return queries.UpsertGeneralConfig(context.Background(), newGeneralConfigParams(config))
}
