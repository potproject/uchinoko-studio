-- name: GetGeneralConfig :one
SELECT *
FROM general_config
WHERE id = 1;

-- name: UpsertGeneralConfig :exec
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
    sqlc.arg(id),
    sqlc.arg(background),
    sqlc.arg(language),
    sqlc.arg(sound_effect),
    sqlc.arg(character_output_change),
    sqlc.arg(enable_tts_optimization),
    sqlc.arg(transcription_type),
    sqlc.arg(transcription_method),
    sqlc.arg(transcription_auto_threshold),
    sqlc.arg(transcription_auto_silent_threshold),
    sqlc.arg(transcription_auto_audio_min_length)
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
    transcription_auto_audio_min_length = excluded.transcription_auto_audio_min_length;

-- name: ListGeneralConfigs :many
SELECT *
FROM general_config
ORDER BY id;
