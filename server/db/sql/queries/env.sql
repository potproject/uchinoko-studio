-- name: GetEnvConfig :one
SELECT *
FROM env_config
WHERE id = 1;

-- name: UpsertEnvConfig :exec
INSERT INTO env_config (
    id,
    openai_speech_to_text_api_key,
    google_speech_to_text_api_key,
    vosk_server_endpoint,
    openai_api_key,
    anthropic_api_key,
    deepseek_api_key,
    gemini_api_key,
    openai_local_api_key,
    openai_local_api_endpoint,
    voicevox_endpoint,
    bertvits2_endpoint,
    irodori_tts_endpoint,
    nijivoice_api_key,
    stylebertvit2_endpoint,
    google_text_to_speech_api_key,
    openai_speech_api_key
) VALUES (
    sqlc.arg(id),
    sqlc.arg(openai_speech_to_text_api_key),
    sqlc.arg(google_speech_to_text_api_key),
    sqlc.arg(vosk_server_endpoint),
    sqlc.arg(openai_api_key),
    sqlc.arg(anthropic_api_key),
    sqlc.arg(deepseek_api_key),
    sqlc.arg(gemini_api_key),
    sqlc.arg(openai_local_api_key),
    sqlc.arg(openai_local_api_endpoint),
    sqlc.arg(voicevox_endpoint),
    sqlc.arg(bertvits2_endpoint),
    sqlc.arg(irodori_tts_endpoint),
    sqlc.arg(nijivoice_api_key),
    sqlc.arg(stylebertvit2_endpoint),
    sqlc.arg(google_text_to_speech_api_key),
    sqlc.arg(openai_speech_api_key)
)
ON CONFLICT(id) DO UPDATE SET
    openai_speech_to_text_api_key = excluded.openai_speech_to_text_api_key,
    google_speech_to_text_api_key = excluded.google_speech_to_text_api_key,
    vosk_server_endpoint = excluded.vosk_server_endpoint,
    openai_api_key = excluded.openai_api_key,
    anthropic_api_key = excluded.anthropic_api_key,
    deepseek_api_key = excluded.deepseek_api_key,
    gemini_api_key = excluded.gemini_api_key,
    openai_local_api_key = excluded.openai_local_api_key,
    openai_local_api_endpoint = excluded.openai_local_api_endpoint,
    voicevox_endpoint = excluded.voicevox_endpoint,
    bertvits2_endpoint = excluded.bertvits2_endpoint,
    irodori_tts_endpoint = excluded.irodori_tts_endpoint,
    nijivoice_api_key = excluded.nijivoice_api_key,
    stylebertvit2_endpoint = excluded.stylebertvit2_endpoint,
    google_text_to_speech_api_key = excluded.google_text_to_speech_api_key,
    openai_speech_api_key = excluded.openai_speech_api_key;

-- name: ListEnvConfigs :many
SELECT *
FROM env_config
ORDER BY id;
