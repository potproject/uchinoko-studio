-- name: ListCharacters :many
SELECT *
FROM characters
ORDER BY name, id;

-- name: GetCharacter :one
SELECT *
FROM characters
WHERE id = sqlc.arg(id);

-- name: UpsertCharacter :exec
INSERT INTO characters (
    id,
    name,
    multi_voice
) VALUES (
    sqlc.arg(id),
    sqlc.arg(name),
    sqlc.arg(multi_voice)
)
ON CONFLICT(id) DO UPDATE SET
    name = excluded.name,
    multi_voice = excluded.multi_voice;

-- name: DeleteCharacter :exec
DELETE FROM characters
WHERE id = sqlc.arg(id);

-- name: GetCharacterChatSetting :one
SELECT *
FROM character_chat_settings
WHERE character_id = sqlc.arg(character_id);

-- name: UpsertCharacterChatSetting :exec
INSERT INTO character_chat_settings (
    character_id,
    type,
    model,
    system_prompt,
    temperature_enable,
    temperature_value,
    max_history
) VALUES (
    sqlc.arg(character_id),
    sqlc.arg(type),
    sqlc.arg(model),
    sqlc.arg(system_prompt),
    sqlc.arg(temperature_enable),
    sqlc.arg(temperature_value),
    sqlc.arg(max_history)
)
ON CONFLICT(character_id) DO UPDATE SET
    type = excluded.type,
    model = excluded.model,
    system_prompt = excluded.system_prompt,
    temperature_enable = excluded.temperature_enable,
    temperature_value = excluded.temperature_value,
    max_history = excluded.max_history;

-- name: DeleteCharacterChatLimits :exec
DELETE FROM character_chat_limits
WHERE character_id = sqlc.arg(character_id);

-- name: InsertCharacterChatLimit :exec
INSERT INTO character_chat_limits (
    character_id,
    window,
    request_limit,
    token_limit
) VALUES (
    sqlc.arg(character_id),
    sqlc.arg(window),
    sqlc.arg(request_limit),
    sqlc.arg(token_limit)
);

-- name: ListCharacterChatLimits :many
SELECT *
FROM character_chat_limits
WHERE character_id = sqlc.arg(character_id)
ORDER BY CASE window
    WHEN 'day' THEN 1
    WHEN 'hour' THEN 2
    WHEN 'minute' THEN 3
    ELSE 4
END;

-- name: DeleteCharacterVoices :exec
DELETE FROM character_voices
WHERE character_id = sqlc.arg(character_id);

-- name: InsertCharacterVoice :exec
INSERT INTO character_voices (
    character_id,
    voice_index,
    name,
    type,
    identification,
    model_id,
    model_file,
    speaker_id,
    reference_audio_path,
    image,
    background_image_path
) VALUES (
    sqlc.arg(character_id),
    sqlc.arg(voice_index),
    sqlc.arg(name),
    sqlc.arg(type),
    sqlc.arg(identification),
    sqlc.arg(model_id),
    sqlc.arg(model_file),
    sqlc.arg(speaker_id),
    sqlc.arg(reference_audio_path),
    sqlc.arg(image),
    sqlc.arg(background_image_path)
);

-- name: ListCharacterVoices :many
SELECT *
FROM character_voices
WHERE character_id = sqlc.arg(character_id)
ORDER BY voice_index;

-- name: InsertCharacterVoiceBehavior :exec
INSERT INTO character_voice_behaviors (
    character_id,
    voice_index,
    behavior_index,
    identification,
    image_path
) VALUES (
    sqlc.arg(character_id),
    sqlc.arg(voice_index),
    sqlc.arg(behavior_index),
    sqlc.arg(identification),
    sqlc.arg(image_path)
);

-- name: ListCharacterVoiceBehaviors :many
SELECT *
FROM character_voice_behaviors
WHERE character_id = sqlc.arg(character_id)
ORDER BY voice_index, behavior_index;
