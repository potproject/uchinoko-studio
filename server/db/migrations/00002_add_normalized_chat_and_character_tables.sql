-- +goose Up
CREATE TABLE IF NOT EXISTS character_chat_settings (
    character_id TEXT PRIMARY KEY,
    type TEXT NOT NULL,
    model TEXT NOT NULL,
    system_prompt TEXT NOT NULL,
    temperature_enable INTEGER NOT NULL,
    temperature_value REAL NOT NULL,
    max_history INTEGER NOT NULL,
    FOREIGN KEY (character_id) REFERENCES characters(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS character_chat_limits (
    character_id TEXT NOT NULL,
    window TEXT NOT NULL CHECK (window IN ('day', 'hour', 'minute')),
    request_limit INTEGER NOT NULL,
    token_limit INTEGER NOT NULL,
    PRIMARY KEY (character_id, window),
    FOREIGN KEY (character_id) REFERENCES characters(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS character_voices (
    character_id TEXT NOT NULL,
    voice_index INTEGER NOT NULL,
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    identification TEXT NOT NULL,
    model_id TEXT NOT NULL,
    model_file TEXT NOT NULL,
    speaker_id TEXT NOT NULL,
    reference_audio_path TEXT NOT NULL,
    image TEXT NOT NULL,
    background_image_path TEXT NOT NULL,
    PRIMARY KEY (character_id, voice_index),
    FOREIGN KEY (character_id) REFERENCES characters(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS character_voice_behaviors (
    character_id TEXT NOT NULL,
    voice_index INTEGER NOT NULL,
    behavior_index INTEGER NOT NULL,
    identification TEXT NOT NULL,
    image_path TEXT NOT NULL,
    PRIMARY KEY (character_id, voice_index, behavior_index),
    FOREIGN KEY (character_id, voice_index) REFERENCES character_voices(character_id, voice_index) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS chat_messages (
    session_id TEXT NOT NULL,
    character_id TEXT NOT NULL,
    message_index INTEGER NOT NULL,
    role TEXT NOT NULL,
    content TEXT NOT NULL,
    image_extension TEXT NOT NULL,
    image_data BLOB,
    PRIMARY KEY (session_id, character_id, message_index),
    FOREIGN KEY (session_id, character_id) REFERENCES chat_sessions(session_id, character_id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE IF EXISTS chat_messages;
DROP TABLE IF EXISTS character_voice_behaviors;
DROP TABLE IF EXISTS character_voices;
DROP TABLE IF EXISTS character_chat_limits;
DROP TABLE IF EXISTS character_chat_settings;
