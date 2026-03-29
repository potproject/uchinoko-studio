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
    multi_voice,
    voice_json,
    chat_json
) VALUES (
    sqlc.arg(id),
    sqlc.arg(name),
    sqlc.arg(multi_voice),
    sqlc.arg(voice_json),
    sqlc.arg(chat_json)
)
ON CONFLICT(id) DO UPDATE SET
    name = excluded.name,
    multi_voice = excluded.multi_voice,
    voice_json = excluded.voice_json,
    chat_json = excluded.chat_json;

-- name: DeleteCharacter :exec
DELETE FROM characters
WHERE id = sqlc.arg(id);
