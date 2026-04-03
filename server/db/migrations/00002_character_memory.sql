-- +goose Up
CREATE TABLE IF NOT EXISTS character_memory_settings (
    character_id TEXT PRIMARY KEY,
    enabled INTEGER NOT NULL,
    max_items_in_prompt INTEGER NOT NULL,
    enable_relationship_memory INTEGER NOT NULL,
    enable_session_summary INTEGER NOT NULL,
    enable_semantic_search INTEGER NOT NULL,
    embedding_model TEXT NOT NULL,
    allow_sensitive_memory INTEGER NOT NULL,
    FOREIGN KEY (character_id) REFERENCES characters(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS memory_items (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    public_id TEXT NOT NULL UNIQUE,
    character_id TEXT NOT NULL,
    owner_id TEXT NOT NULL DEFAULT '',
    scope TEXT NOT NULL CHECK (scope IN ('character', 'relationship')),
    kind TEXT NOT NULL,
    content TEXT NOT NULL,
    normalized_content TEXT NOT NULL,
    keywords_text TEXT NOT NULL,
    pinned INTEGER NOT NULL DEFAULT 0,
    confidence REAL NOT NULL DEFAULT 0,
    salience REAL NOT NULL DEFAULT 0,
    source TEXT NOT NULL DEFAULT 'manual',
    superseded_by INTEGER,
    embedding_json TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    last_accessed_at TEXT NOT NULL,
    FOREIGN KEY (character_id) REFERENCES characters(id) ON DELETE CASCADE,
    FOREIGN KEY (superseded_by) REFERENCES memory_items(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_memory_items_character_owner_scope ON memory_items(character_id, owner_id, scope);
CREATE INDEX IF NOT EXISTS idx_memory_items_public_id ON memory_items(public_id);

CREATE VIRTUAL TABLE IF NOT EXISTS memory_items_fts USING fts5(
    normalized_content,
    keywords_text,
    content='memory_items',
    content_rowid='id',
    tokenize='trigram'
);

-- +goose StatementBegin
CREATE TRIGGER IF NOT EXISTS memory_items_ai AFTER INSERT ON memory_items BEGIN
    INSERT INTO memory_items_fts(rowid, normalized_content, keywords_text)
    VALUES (new.id, new.normalized_content, new.keywords_text);
END;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER IF NOT EXISTS memory_items_ad AFTER DELETE ON memory_items BEGIN
    INSERT INTO memory_items_fts(memory_items_fts, rowid, normalized_content, keywords_text)
    VALUES('delete', old.id, old.normalized_content, old.keywords_text);
END;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER IF NOT EXISTS memory_items_au AFTER UPDATE ON memory_items BEGIN
    INSERT INTO memory_items_fts(memory_items_fts, rowid, normalized_content, keywords_text)
    VALUES('delete', old.id, old.normalized_content, old.keywords_text);
    INSERT INTO memory_items_fts(rowid, normalized_content, keywords_text)
    VALUES (new.id, new.normalized_content, new.keywords_text);
END;
-- +goose StatementEnd

CREATE TABLE IF NOT EXISTS memory_item_evidence (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    memory_item_id INTEGER NOT NULL,
    session_id TEXT NOT NULL,
    message_index_start INTEGER NOT NULL,
    message_index_end INTEGER NOT NULL,
    excerpt TEXT NOT NULL,
    created_at TEXT NOT NULL,
    FOREIGN KEY (memory_item_id) REFERENCES memory_items(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_memory_item_evidence_item ON memory_item_evidence(memory_item_id);

CREATE TABLE IF NOT EXISTS session_summaries (
    owner_id TEXT NOT NULL,
    character_id TEXT NOT NULL,
    session_id TEXT NOT NULL,
    summary TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    PRIMARY KEY (owner_id, character_id, session_id),
    FOREIGN KEY (character_id) REFERENCES characters(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS memory_jobs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    type TEXT NOT NULL CHECK (type IN ('extract_turn', 'compact_session', 'reindex_memory_item', 'rebuild_character_memory')),
    status TEXT NOT NULL CHECK (status IN ('queued', 'running', 'failed')),
    unique_key TEXT NOT NULL UNIQUE,
    payload_json TEXT NOT NULL,
    attempts INTEGER NOT NULL DEFAULT 0,
    available_at TEXT NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    last_error TEXT NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_memory_jobs_ready ON memory_jobs(status, available_at, id);

INSERT INTO character_memory_settings (
    character_id,
    enabled,
    max_items_in_prompt,
    enable_relationship_memory,
    enable_session_summary,
    enable_semantic_search,
    embedding_model,
    allow_sensitive_memory
)
SELECT
    id,
    0,
    6,
    1,
    1,
    1,
    'text-embedding-3-small',
    0
FROM characters
WHERE NOT EXISTS (
    SELECT 1 FROM character_memory_settings cms WHERE cms.character_id = characters.id
);

-- +goose Down
DROP TABLE IF EXISTS memory_jobs;
DROP TABLE IF EXISTS session_summaries;
DROP TABLE IF EXISTS memory_item_evidence;
DROP TRIGGER IF EXISTS memory_items_au;
DROP TRIGGER IF EXISTS memory_items_ad;
DROP TRIGGER IF EXISTS memory_items_ai;
DROP TABLE IF EXISTS memory_items_fts;
DROP TABLE IF EXISTS memory_items;
DROP TABLE IF EXISTS character_memory_settings;
