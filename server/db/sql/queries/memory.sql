-- name: GetCharacterMemorySetting :one
SELECT
    character_id,
    enabled,
    max_items_in_prompt,
    enable_relationship_memory,
    enable_session_summary,
    enable_semantic_search,
    embedding_model,
    allow_sensitive_memory
FROM character_memory_settings
WHERE character_id = sqlc.arg(character_id);

-- name: UpsertCharacterMemorySetting :exec
INSERT INTO character_memory_settings (
    character_id,
    enabled,
    max_items_in_prompt,
    enable_relationship_memory,
    enable_session_summary,
    enable_semantic_search,
    embedding_model,
    allow_sensitive_memory
) VALUES (
    sqlc.arg(character_id),
    sqlc.arg(enabled),
    sqlc.arg(max_items_in_prompt),
    sqlc.arg(enable_relationship_memory),
    sqlc.arg(enable_session_summary),
    sqlc.arg(enable_semantic_search),
    sqlc.arg(embedding_model),
    sqlc.arg(allow_sensitive_memory)
)
ON CONFLICT(character_id) DO UPDATE SET
    enabled = excluded.enabled,
    max_items_in_prompt = excluded.max_items_in_prompt,
    enable_relationship_memory = excluded.enable_relationship_memory,
    enable_session_summary = excluded.enable_session_summary,
    enable_semantic_search = excluded.enable_semantic_search,
    embedding_model = excluded.embedding_model,
    allow_sensitive_memory = excluded.allow_sensitive_memory;

-- name: ListMemoryItemsForOwnerAndCharacter :many
SELECT
    id,
    public_id,
    character_id,
    owner_id,
    scope,
    kind,
    content,
    normalized_content,
    keywords_text,
    pinned,
    confidence,
    salience,
    source,
    superseded_by,
    embedding_json,
    created_at,
    updated_at,
    last_accessed_at
FROM memory_items
WHERE character_id = sqlc.arg(character_id)
  AND superseded_by IS NULL
  AND (
    (scope = 'character' AND owner_id = '')
    OR (scope = 'relationship' AND owner_id = sqlc.arg(owner_id))
  )
ORDER BY pinned DESC, updated_at DESC, id DESC;

-- name: InsertMemoryItem :execresult
INSERT INTO memory_items (
    public_id,
    character_id,
    owner_id,
    scope,
    kind,
    content,
    normalized_content,
    keywords_text,
    pinned,
    confidence,
    salience,
    source,
    superseded_by,
    embedding_json,
    created_at,
    updated_at,
    last_accessed_at
) VALUES (
    sqlc.arg(public_id),
    sqlc.arg(character_id),
    sqlc.arg(owner_id),
    sqlc.arg(scope),
    sqlc.arg(kind),
    sqlc.arg(content),
    sqlc.arg(normalized_content),
    sqlc.arg(keywords_text),
    sqlc.arg(pinned),
    sqlc.arg(confidence),
    sqlc.arg(salience),
    sqlc.arg(source),
    sqlc.arg(superseded_by),
    sqlc.arg(embedding_json),
    sqlc.arg(created_at),
    sqlc.arg(updated_at),
    sqlc.arg(last_accessed_at)
);

-- name: UpdateMemoryItemEditable :exec
UPDATE memory_items
SET kind = sqlc.arg(kind),
    content = sqlc.arg(content),
    normalized_content = sqlc.arg(normalized_content),
    keywords_text = sqlc.arg(keywords_text),
    pinned = sqlc.arg(pinned),
    confidence = sqlc.arg(confidence),
    salience = sqlc.arg(salience),
    updated_at = sqlc.arg(updated_at),
    last_accessed_at = sqlc.arg(last_accessed_at)
WHERE id = sqlc.arg(id);

-- name: DeleteMemoryItemByPublicID :exec
DELETE FROM memory_items
WHERE public_id = sqlc.arg(public_id);

-- name: GetSessionSummary :one
SELECT
    owner_id,
    character_id,
    session_id,
    summary,
    updated_at
FROM session_summaries
WHERE owner_id = sqlc.arg(owner_id)
  AND character_id = sqlc.arg(character_id)
  AND session_id = sqlc.arg(session_id);

-- name: UpsertSessionSummary :exec
INSERT INTO session_summaries (
    owner_id,
    character_id,
    session_id,
    summary,
    updated_at
) VALUES (
    sqlc.arg(owner_id),
    sqlc.arg(character_id),
    sqlc.arg(session_id),
    sqlc.arg(summary),
    sqlc.arg(updated_at)
)
ON CONFLICT(owner_id, character_id, session_id) DO UPDATE SET
    summary = excluded.summary,
    updated_at = excluded.updated_at;

-- name: GetMemoryItemByPublicID :one
SELECT
    id,
    public_id,
    character_id,
    owner_id,
    scope,
    kind,
    content,
    normalized_content,
    keywords_text,
    pinned,
    confidence,
    salience,
    source,
    superseded_by,
    embedding_json,
    created_at,
    updated_at,
    last_accessed_at
FROM memory_items
WHERE public_id = sqlc.arg(public_id);

-- name: SearchMemoryFallbackWithRelationship :many
SELECT
    id,
    public_id,
    character_id,
    owner_id,
    scope,
    kind,
    content,
    normalized_content,
    keywords_text,
    pinned,
    confidence,
    salience,
    source,
    superseded_by,
    embedding_json,
    created_at,
    updated_at,
    last_accessed_at,
    0.0 AS lexical_raw
FROM memory_items
WHERE character_id = sqlc.arg(character_id)
  AND superseded_by IS NULL
  AND (
    (scope = 'character' AND owner_id = '')
    OR (scope = 'relationship' AND owner_id = sqlc.arg(owner_id))
  )
  AND (
    normalized_content LIKE '%' || sqlc.arg(query) || '%'
    OR keywords_text LIKE '%' || sqlc.arg(query) || '%'
  )
ORDER BY pinned DESC, updated_at DESC
LIMIT sqlc.arg(limit_rows);

-- name: SearchMemoryFallbackCharacterOnly :many
SELECT
    id,
    public_id,
    character_id,
    owner_id,
    scope,
    kind,
    content,
    normalized_content,
    keywords_text,
    pinned,
    confidence,
    salience,
    source,
    superseded_by,
    embedding_json,
    created_at,
    updated_at,
    last_accessed_at,
    0.0 AS lexical_raw
FROM memory_items
WHERE character_id = sqlc.arg(character_id)
  AND superseded_by IS NULL
  AND scope = 'character'
  AND owner_id = ''
  AND (
    normalized_content LIKE '%' || sqlc.arg(query) || '%'
    OR keywords_text LIKE '%' || sqlc.arg(query) || '%'
  )
ORDER BY pinned DESC, updated_at DESC
LIMIT sqlc.arg(limit_rows);

-- name: ListPinnedMemoryCandidatesWithRelationship :many
SELECT
    id,
    public_id,
    character_id,
    owner_id,
    scope,
    kind,
    content,
    normalized_content,
    keywords_text,
    pinned,
    confidence,
    salience,
    source,
    superseded_by,
    embedding_json,
    created_at,
    updated_at,
    last_accessed_at
FROM memory_items
WHERE character_id = sqlc.arg(character_id)
  AND superseded_by IS NULL
  AND pinned = 1
  AND (
    (scope = 'character' AND owner_id = '')
    OR (scope = 'relationship' AND owner_id = sqlc.arg(owner_id))
  )
ORDER BY updated_at DESC
LIMIT sqlc.arg(limit_rows);

-- name: ListPinnedMemoryCandidatesCharacterOnly :many
SELECT
    id,
    public_id,
    character_id,
    owner_id,
    scope,
    kind,
    content,
    normalized_content,
    keywords_text,
    pinned,
    confidence,
    salience,
    source,
    superseded_by,
    embedding_json,
    created_at,
    updated_at,
    last_accessed_at
FROM memory_items
WHERE character_id = sqlc.arg(character_id)
  AND superseded_by IS NULL
  AND pinned = 1
  AND scope = 'character'
  AND owner_id = ''
ORDER BY updated_at DESC
LIMIT sqlc.arg(limit_rows);

-- name: InsertMemoryJobIgnoreConflict :exec
INSERT INTO memory_jobs (
    type,
    status,
    unique_key,
    payload_json,
    attempts,
    available_at,
    created_at,
    updated_at,
    last_error
) VALUES (
    sqlc.arg(type),
    sqlc.arg(status),
    sqlc.arg(unique_key),
    sqlc.arg(payload_json),
    sqlc.arg(attempts),
    sqlc.arg(available_at),
    sqlc.arg(created_at),
    sqlc.arg(updated_at),
    sqlc.arg(last_error)
)
ON CONFLICT(unique_key) DO NOTHING;

-- name: RecoverRunningMemoryJobs :exec
UPDATE memory_jobs
SET status = sqlc.arg(status),
    updated_at = sqlc.arg(updated_at)
WHERE status = sqlc.arg(previous_status);

-- name: GetNextQueuedMemoryJob :one
SELECT
    id,
    type,
    status,
    unique_key,
    payload_json,
    attempts,
    available_at,
    created_at,
    updated_at,
    last_error
FROM memory_jobs
WHERE status = sqlc.arg(status)
  AND available_at <= sqlc.arg(available_at)
ORDER BY available_at ASC, id ASC
LIMIT 1;

-- name: MarkMemoryJobRunning :exec
UPDATE memory_jobs
SET status = sqlc.arg(status),
    attempts = attempts + 1,
    updated_at = sqlc.arg(updated_at)
WHERE id = sqlc.arg(id);

-- name: DeleteMemoryJob :exec
DELETE FROM memory_jobs
WHERE id = sqlc.arg(id);

-- name: CountMemoryJobs :one
SELECT COUNT(*)
FROM memory_jobs;

-- name: UpdateMemoryJobFailure :exec
UPDATE memory_jobs
SET status = sqlc.arg(status),
    available_at = sqlc.arg(available_at),
    updated_at = sqlc.arg(updated_at),
    last_error = sqlc.arg(last_error)
WHERE id = sqlc.arg(id);

-- name: ListMemoryRecordsForCharacter :many
SELECT
    id,
    public_id,
    character_id,
    owner_id,
    scope,
    kind,
    content,
    normalized_content,
    keywords_text,
    pinned,
    confidence,
    salience,
    source,
    superseded_by,
    embedding_json,
    created_at,
    updated_at,
    last_accessed_at
FROM memory_items
WHERE character_id = sqlc.arg(character_id)
  AND superseded_by IS NULL
  AND (
    owner_id = ''
    OR owner_id = sqlc.arg(owner_id)
  )
ORDER BY id;

-- name: UpdateMemoryItemIndex :exec
UPDATE memory_items
SET normalized_content = sqlc.arg(normalized_content),
    keywords_text = sqlc.arg(keywords_text),
    embedding_json = sqlc.arg(embedding_json),
    updated_at = sqlc.arg(updated_at)
WHERE id = sqlc.arg(id);

-- name: ListActiveRelationshipMemory :many
SELECT
    id,
    public_id,
    character_id,
    owner_id,
    scope,
    kind,
    content,
    normalized_content,
    keywords_text,
    pinned,
    confidence,
    salience,
    source,
    superseded_by,
    embedding_json,
    created_at,
    updated_at,
    last_accessed_at
FROM memory_items
WHERE character_id = sqlc.arg(character_id)
  AND owner_id = sqlc.arg(owner_id)
  AND scope = 'relationship'
  AND superseded_by IS NULL;

-- name: SupersedeMemoryItem :exec
UPDATE memory_items
SET superseded_by = sqlc.arg(superseded_by)
WHERE id = sqlc.arg(id);

-- name: InsertMemoryItemEvidence :exec
INSERT INTO memory_item_evidence (
    memory_item_id,
    session_id,
    message_index_start,
    message_index_end,
    excerpt,
    created_at
) VALUES (
    sqlc.arg(memory_item_id),
    sqlc.arg(session_id),
    sqlc.arg(message_index_start),
    sqlc.arg(message_index_end),
    sqlc.arg(excerpt),
    sqlc.arg(created_at)
);
