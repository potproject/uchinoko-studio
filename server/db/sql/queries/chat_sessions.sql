-- name: GetChatSession :one
SELECT *
FROM chat_sessions
WHERE session_id = sqlc.arg(session_id)
  AND character_id = sqlc.arg(character_id);

-- name: UpsertChatSession :exec
INSERT INTO chat_sessions (
    session_id,
    character_id
) VALUES (
    sqlc.arg(session_id),
    sqlc.arg(character_id)
)
ON CONFLICT(session_id, character_id) DO NOTHING;

-- name: DeleteChatSession :exec
DELETE FROM chat_sessions
WHERE session_id = sqlc.arg(session_id)
  AND character_id = sqlc.arg(character_id);

-- name: ListChatSessions :many
SELECT *
FROM chat_sessions
ORDER BY session_id, character_id;

-- name: ListChatSessionsByCharacter :many
SELECT *
FROM chat_sessions
WHERE character_id = sqlc.arg(character_id)
ORDER BY session_id;

-- name: DeleteChatMessages :exec
DELETE FROM chat_messages
WHERE session_id = sqlc.arg(session_id)
  AND character_id = sqlc.arg(character_id);

-- name: InsertChatMessage :exec
INSERT INTO chat_messages (
    session_id,
    character_id,
    message_index,
    role,
    content,
    image_extension,
    image_data
) VALUES (
    sqlc.arg(session_id),
    sqlc.arg(character_id),
    sqlc.arg(message_index),
    sqlc.arg(role),
    sqlc.arg(content),
    sqlc.arg(image_extension),
    sqlc.arg(image_data)
);

-- name: ListChatMessages :many
SELECT *
FROM chat_messages
WHERE session_id = sqlc.arg(session_id)
  AND character_id = sqlc.arg(character_id)
ORDER BY message_index;
