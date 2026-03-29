-- name: GetChatSession :one
SELECT *
FROM chat_sessions
WHERE session_id = sqlc.arg(session_id)
  AND character_id = sqlc.arg(character_id);

-- name: UpsertChatSession :exec
INSERT INTO chat_sessions (
    session_id,
    character_id,
    messages_json
) VALUES (
    sqlc.arg(session_id),
    sqlc.arg(character_id),
    sqlc.arg(messages_json)
)
ON CONFLICT(session_id, character_id) DO UPDATE SET
    messages_json = excluded.messages_json;

-- name: DeleteChatSession :exec
DELETE FROM chat_sessions
WHERE session_id = sqlc.arg(session_id)
  AND character_id = sqlc.arg(character_id);

-- name: ListChatSessions :many
SELECT *
FROM chat_sessions
ORDER BY session_id, character_id;
