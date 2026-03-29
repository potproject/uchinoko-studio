package db

import "github.com/potproject/uchinoko-studio/data"

type generalConfigRow struct {
	ID                               int     `db:"id"`
	Background                       string  `db:"background"`
	Language                         string  `db:"language"`
	SoundEffect                      int     `db:"sound_effect"`
	CharacterOutputChange            int     `db:"character_output_change"`
	EnableTTSOptimization            int     `db:"enable_tts_optimization"`
	TranscriptionType                string  `db:"transcription_type"`
	TranscriptionMethod              string  `db:"transcription_method"`
	TranscriptionAutoThreshold       float64 `db:"transcription_auto_threshold"`
	TranscriptionAutoSilentThreshold float64 `db:"transcription_auto_silent_threshold"`
	TranscriptionAutoAudioMinLength  float64 `db:"transcription_auto_audio_min_length"`
}

func (r generalConfigRow) toConfig() data.GeneralConfig {
	return data.GeneralConfig{
		Background:            r.Background,
		Language:              r.Language,
		SoundEffect:           intToBool(r.SoundEffect),
		CharacterOutputChange: intToBool(r.CharacterOutputChange),
		EnableTTSOptimization: intToBool(r.EnableTTSOptimization),
		Transcription: struct {
			Type        string `json:"type"`
			Method      string `json:"method"`
			AutoSetting struct {
				Threshold       float64 `json:"threshold"`
				SilentThreshold float64 `json:"silentThreshold"`
				AudioMinLength  float64 `json:"audioMinLength"`
			} `json:"autoSetting"`
		}{
			Type:   r.TranscriptionType,
			Method: r.TranscriptionMethod,
			AutoSetting: struct {
				Threshold       float64 `json:"threshold"`
				SilentThreshold float64 `json:"silentThreshold"`
				AudioMinLength  float64 `json:"audioMinLength"`
			}{
				Threshold:       r.TranscriptionAutoThreshold,
				SilentThreshold: r.TranscriptionAutoSilentThreshold,
				AudioMinLength:  r.TranscriptionAutoAudioMinLength,
			},
		},
	}
}

func newGeneralConfigRow(config data.GeneralConfig) generalConfigRow {
	return generalConfigRow{
		ID:                               1,
		Background:                       config.Background,
		Language:                         config.Language,
		SoundEffect:                      boolToInt(config.SoundEffect),
		CharacterOutputChange:            boolToInt(config.CharacterOutputChange),
		EnableTTSOptimization:            boolToInt(config.EnableTTSOptimization),
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
	var row generalConfigRow
	err := db.Get(&row, "SELECT * FROM general_config WHERE id = 1")
	if isNotFound(err) {
		return generalInitConfig(), nil
	}
	if err != nil {
		return data.GeneralConfig{}, err
	}

	return row.toConfig(), nil
}

func PutGeneralConfig(config data.GeneralConfig) error {
	row := newGeneralConfigRow(config)

	_, err := db.NamedExec(`
		INSERT INTO general_config (
			id,
			background,
			language,
			sound_effect,
			character_output_change,
			enable_tts_optimization,
			transcription_type,
			transcription_method,
			transcription_auto_threshold,
			transcription_auto_silent_threshold,
			transcription_auto_audio_min_length
		) VALUES (
			:id,
			:background,
			:language,
			:sound_effect,
			:character_output_change,
			:enable_tts_optimization,
			:transcription_type,
			:transcription_method,
			:transcription_auto_threshold,
			:transcription_auto_silent_threshold,
			:transcription_auto_audio_min_length
		)
		ON CONFLICT(id) DO UPDATE SET
			background = excluded.background,
			language = excluded.language,
			sound_effect = excluded.sound_effect,
			character_output_change = excluded.character_output_change,
			enable_tts_optimization = excluded.enable_tts_optimization,
			transcription_type = excluded.transcription_type,
			transcription_method = excluded.transcription_method,
			transcription_auto_threshold = excluded.transcription_auto_threshold,
			transcription_auto_silent_threshold = excluded.transcription_auto_silent_threshold,
			transcription_auto_audio_min_length = excluded.transcription_auto_audio_min_length
	`, row)
	return err
}
