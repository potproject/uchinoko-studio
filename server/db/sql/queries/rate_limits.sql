-- name: GetRateLimit :one
SELECT *
FROM rate_limits
WHERE id = sqlc.arg(id);

-- name: UpsertRateLimit :exec
INSERT INTO rate_limits (
    id,
    day_last_update,
    day_request,
    day_token,
    hour_last_update,
    hour_request,
    hour_token,
    minute_last_update,
    minute_request,
    minute_token
) VALUES (
    sqlc.arg(id),
    sqlc.arg(day_last_update),
    sqlc.arg(day_request),
    sqlc.arg(day_token),
    sqlc.arg(hour_last_update),
    sqlc.arg(hour_request),
    sqlc.arg(hour_token),
    sqlc.arg(minute_last_update),
    sqlc.arg(minute_request),
    sqlc.arg(minute_token)
)
ON CONFLICT(id) DO UPDATE SET
    day_last_update = excluded.day_last_update,
    day_request = excluded.day_request,
    day_token = excluded.day_token,
    hour_last_update = excluded.hour_last_update,
    hour_request = excluded.hour_request,
    hour_token = excluded.hour_token,
    minute_last_update = excluded.minute_last_update,
    minute_request = excluded.minute_request,
    minute_token = excluded.minute_token;

-- name: ListRateLimits :many
SELECT *
FROM rate_limits
ORDER BY id;
