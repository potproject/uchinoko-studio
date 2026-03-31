-- +goose Up
CREATE TABLE IF NOT EXISTS general_config (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    background TEXT NOT NULL,
    language TEXT NOT NULL,
    sound_effect INTEGER NOT NULL,
    character_output_change INTEGER NOT NULL,
    enable_tts_optimization INTEGER NOT NULL,
    transcription_type TEXT NOT NULL,
    transcription_method TEXT NOT NULL,
    transcription_auto_threshold REAL NOT NULL,
    transcription_auto_silent_threshold REAL NOT NULL,
    transcription_auto_audio_min_length REAL NOT NULL
);

CREATE TABLE IF NOT EXISTS env_config (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    openai_speech_to_text_api_key TEXT NOT NULL,
    google_speech_to_text_api_key TEXT NOT NULL,
    vosk_server_endpoint TEXT NOT NULL,
    openai_api_key TEXT NOT NULL,
    anthropic_api_key TEXT NOT NULL,
    deepseek_api_key TEXT NOT NULL,
    gemini_api_key TEXT NOT NULL,
    openai_local_api_key TEXT NOT NULL,
    openai_local_api_endpoint TEXT NOT NULL,
    voicevox_endpoint TEXT NOT NULL,
    bertvits2_endpoint TEXT NOT NULL,
    irodori_tts_endpoint TEXT NOT NULL,
    nijivoice_api_key TEXT NOT NULL,
    stylebertvit2_endpoint TEXT NOT NULL,
    google_text_to_speech_api_key TEXT NOT NULL,
    openai_speech_api_key TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS characters (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    multi_voice INTEGER NOT NULL
);

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

CREATE TABLE IF NOT EXISTS chat_sessions (
    session_id TEXT NOT NULL,
    character_id TEXT NOT NULL,
    PRIMARY KEY (session_id, character_id)
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

CREATE TABLE IF NOT EXISTS rate_limits (
    id TEXT PRIMARY KEY,
    day_last_update TEXT NOT NULL,
    day_request INTEGER NOT NULL,
    day_token INTEGER NOT NULL,
    hour_last_update TEXT NOT NULL,
    hour_request INTEGER NOT NULL,
    hour_token INTEGER NOT NULL,
    minute_last_update TEXT NOT NULL,
    minute_request INTEGER NOT NULL,
    minute_token INTEGER NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS rate_limits;
DROP TABLE IF EXISTS chat_messages;
DROP TABLE IF EXISTS chat_sessions;
DROP TABLE IF EXISTS character_voice_behaviors;
DROP TABLE IF EXISTS character_voices;
DROP TABLE IF EXISTS character_chat_limits;
DROP TABLE IF EXISTS character_chat_settings;
DROP TABLE IF EXISTS characters;
DROP TABLE IF EXISTS env_config;
DROP TABLE IF EXISTS general_config;
