-- +goose NO TRANSACTION
-- +goose Up
PRAGMA foreign_keys = OFF;

CREATE TABLE characters_new (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    multi_voice INTEGER NOT NULL
);

INSERT INTO characters_new (
    id,
    name,
    multi_voice
)
SELECT
    id,
    name,
    multi_voice
FROM characters;

DROP TABLE characters;
ALTER TABLE characters_new RENAME TO characters;

CREATE TABLE chat_sessions_new (
    session_id TEXT NOT NULL,
    character_id TEXT NOT NULL,
    PRIMARY KEY (session_id, character_id)
);

INSERT INTO chat_sessions_new (
    session_id,
    character_id
)
SELECT
    session_id,
    character_id
FROM chat_sessions;

DROP TABLE chat_sessions;
ALTER TABLE chat_sessions_new RENAME TO chat_sessions;

PRAGMA foreign_keys = ON;

-- +goose Down
PRAGMA foreign_keys = OFF;

CREATE TABLE characters_old (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    multi_voice INTEGER NOT NULL,
    voice_json TEXT NOT NULL,
    chat_json TEXT NOT NULL
);

INSERT INTO characters_old (
    id,
    name,
    multi_voice,
    voice_json,
    chat_json
)
SELECT
    id,
    name,
    multi_voice,
    '[]',
    '{}'
FROM characters;

DROP TABLE characters;
ALTER TABLE characters_old RENAME TO characters;

CREATE TABLE chat_sessions_old (
    session_id TEXT NOT NULL,
    character_id TEXT NOT NULL,
    messages_json TEXT NOT NULL,
    PRIMARY KEY (session_id, character_id)
);

INSERT INTO chat_sessions_old (
    session_id,
    character_id,
    messages_json
)
SELECT
    session_id,
    character_id,
    '[]'
FROM chat_sessions;

DROP TABLE chat_sessions;
ALTER TABLE chat_sessions_old RENAME TO chat_sessions;

PRAGMA foreign_keys = ON;
