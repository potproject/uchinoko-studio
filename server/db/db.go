package db

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/potproject/uchinoko-studio/envgen"
	_ "modernc.org/sqlite"
)

var db *sqlx.DB

func Start() {
	if err := StartWithPath(envgen.Get().DB_FILE_PATH()); err != nil {
		log.Fatal(err)
	}
}

func StartWithPath(rawPath string) error {
	return startWithConfigPath(rawPath)
}

func Close() error {
	return closeCurrentDB()
}

func startWithConfigPath(rawPath string) error {
	resolvedPath, err := resolveSQLitePath(rawPath)
	if err != nil {
		return err
	}
	return startWithResolvedPath(resolvedPath)
}

func startWithResolvedPath(resolvedPath string) error {
	nextDB, err := openSQLite(resolvedPath)
	if err != nil {
		return err
	}

	if err := closeCurrentDB(); err != nil {
		nextDB.Close()
		return err
	}

	db = nextDB
	return nil
}

func closeCurrentDB() error {
	if db == nil {
		return nil
	}

	if err := db.Close(); err != nil {
		return err
	}

	db = nil
	return nil
}

func openSQLite(resolvedPath string) (*sqlx.DB, error) {
	if resolvedPath == "" {
		return nil, errors.New("sqlite path is empty")
	}

	if err := os.MkdirAll(filepath.Dir(resolvedPath), 0o755); err != nil {
		return nil, fmt.Errorf("create sqlite directory: %w", err)
	}

	conn, err := sqlx.Open("sqlite", resolvedPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	conn.SetMaxOpenConns(1)
	conn.SetMaxIdleConns(1)

	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}

	pragmas := []string{
		"PRAGMA foreign_keys = ON",
		"PRAGMA journal_mode = WAL",
		"PRAGMA busy_timeout = 5000",
	}
	for _, pragma := range pragmas {
		if _, err := conn.Exec(pragma); err != nil {
			conn.Close()
			return nil, fmt.Errorf("apply pragma %q: %w", pragma, err)
		}
	}

	if err := ensureSchema(conn); err != nil {
		conn.Close()
		return nil, err
	}

	return conn, nil
}

func resolveSQLitePath(rawPath string) (string, error) {
	rawPath = strings.TrimSpace(rawPath)
	if rawPath == "" {
		rawPath = "database"
	}

	cleaned := filepath.Clean(rawPath)
	ext := strings.ToLower(filepath.Ext(cleaned))
	if ext == ".db" || ext == ".sqlite" {
		return cleaned, nil
	}

	return filepath.Join(cleaned, "uchinoko.db"), nil
}

func ensureSchema(conn *sqlx.DB) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS general_config (
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
		)`,
		`CREATE TABLE IF NOT EXISTS env_config (
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
		)`,
		`CREATE TABLE IF NOT EXISTS characters (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			multi_voice INTEGER NOT NULL,
			voice_json TEXT NOT NULL,
			chat_json TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS chat_sessions (
			session_id TEXT NOT NULL,
			character_id TEXT NOT NULL,
			messages_json TEXT NOT NULL,
			PRIMARY KEY (session_id, character_id)
		)`,
		`CREATE TABLE IF NOT EXISTS rate_limits (
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
		)`,
	}

	for _, statement := range statements {
		if _, err := conn.Exec(statement); err != nil {
			return fmt.Errorf("ensure schema: %w", err)
		}
	}

	return nil
}

func isNotFound(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func intToBool(value int) bool {
	return value != 0
}

func marshalJSONString(value any) (string, error) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(encoded), nil
}

func mustMarshalJSONString(value any) string {
	encoded, err := marshalJSONString(value)
	if err != nil {
		return "{}"
	}
	return encoded
}

func unmarshalJSONString[T any](encoded string) (T, error) {
	var value T
	if err := json.Unmarshal([]byte(encoded), &value); err != nil {
		return value, err
	}
	return value, nil
}

type ListAllResponse struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func ListAll() ([]ListAllResponse, error) {
	var response []ListAllResponse

	var generalRows []generalConfigRow
	if err := db.Select(&generalRows, "SELECT * FROM general_config ORDER BY id"); err != nil {
		return nil, err
	}
	for _, row := range generalRows {
		response = append(response, ListAllResponse{
			Key:   "general_config",
			Value: mustMarshalJSONString(row.toConfig()),
		})
	}

	var envRows []envConfigRow
	if err := db.Select(&envRows, "SELECT * FROM env_config ORDER BY id"); err != nil {
		return nil, err
	}
	for _, row := range envRows {
		response = append(response, ListAllResponse{
			Key:   "env_config",
			Value: mustMarshalJSONString(row.toConfig()),
		})
	}

	var characterRows []characterRow
	if err := db.Select(&characterRows, "SELECT * FROM characters ORDER BY name, id"); err != nil {
		return nil, err
	}
	for _, row := range characterRows {
		config, err := row.toConfig()
		if err != nil {
			return nil, err
		}
		response = append(response, ListAllResponse{
			Key:   "character_config/" + row.ID,
			Value: mustMarshalJSONString(config),
		})
	}

	var chatRows []chatSessionRow
	if err := db.Select(&chatRows, "SELECT * FROM chat_sessions ORDER BY session_id, character_id"); err != nil {
		return nil, err
	}
	for _, row := range chatRows {
		message, err := row.toMessage()
		if err != nil {
			return nil, err
		}
		response = append(response, ListAllResponse{
			Key:   "chatmessage/" + row.SessionID + "/" + row.CharacterID,
			Value: mustMarshalJSONString(message),
		})
	}

	var rateLimitRows []rateLimitRow
	if err := db.Select(&rateLimitRows, "SELECT * FROM rate_limits ORDER BY id"); err != nil {
		return nil, err
	}
	for _, row := range rateLimitRows {
		response = append(response, ListAllResponse{
			Key:   "rate_limit/" + row.ID,
			Value: mustMarshalJSONString(row.toRateLimit()),
		})
	}

	return response, nil
}
