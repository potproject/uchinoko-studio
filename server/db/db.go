package db

import (
	"context"
	"database/sql"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/potproject/uchinoko-studio/db/sqlcgen"
	"github.com/potproject/uchinoko-studio/envgen"
	_ "modernc.org/sqlite"
)

//go:embed sql/schema.sql
var schemaDDL string

var db *sql.DB
var queries *sqlcgen.Queries

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
	nextQueries := sqlcgen.New(nextDB)

	if err := closeCurrentDB(); err != nil {
		nextDB.Close()
		return err
	}

	db = nextDB
	queries = nextQueries
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
	queries = nil
	return nil
}

func openSQLite(resolvedPath string) (*sql.DB, error) {
	if resolvedPath == "" {
		return nil, errors.New("sqlite path is empty")
	}

	if err := os.MkdirAll(filepath.Dir(resolvedPath), 0o755); err != nil {
		return nil, fmt.Errorf("create sqlite directory: %w", err)
	}

	conn, err := sql.Open("sqlite", resolvedPath)
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

func ensureSchema(conn *sql.DB) error {
	if _, err := conn.Exec(schemaDDL); err != nil {
		return fmt.Errorf("ensure schema: %w", err)
	}

	return nil
}

func isNotFound(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}

func boolToInt(value bool) int64 {
	if value {
		return 1
	}
	return 0
}

func intToBool(value int64) bool {
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
	ctx := context.Background()
	var response []ListAllResponse

	generalRows, err := queries.ListGeneralConfigs(ctx)
	if err != nil {
		return nil, err
	}
	for _, row := range generalRows {
		response = append(response, ListAllResponse{
			Key:   "general_config",
			Value: mustMarshalJSONString(generalConfigFromRow(row)),
		})
	}

	envRows, err := queries.ListEnvConfigs(ctx)
	if err != nil {
		return nil, err
	}
	for _, row := range envRows {
		response = append(response, ListAllResponse{
			Key:   "env_config",
			Value: mustMarshalJSONString(envConfigFromRow(row)),
		})
	}

	characterRows, err := queries.ListCharacters(ctx)
	if err != nil {
		return nil, err
	}
	for _, row := range characterRows {
		config, err := characterConfigFromRow(row)
		if err != nil {
			return nil, err
		}
		response = append(response, ListAllResponse{
			Key:   "character_config/" + row.ID,
			Value: mustMarshalJSONString(config),
		})
	}

	chatRows, err := queries.ListChatSessions(ctx)
	if err != nil {
		return nil, err
	}
	for _, row := range chatRows {
		message, err := chatMessageFromRow(row)
		if err != nil {
			return nil, err
		}
		response = append(response, ListAllResponse{
			Key:   "chatmessage/" + row.SessionID + "/" + row.CharacterID,
			Value: mustMarshalJSONString(message),
		})
	}

	rateLimitRows, err := queries.ListRateLimits(ctx)
	if err != nil {
		return nil, err
	}
	for _, row := range rateLimitRows {
		response = append(response, ListAllResponse{
			Key:   "rate_limit/" + row.ID,
			Value: mustMarshalJSONString(rateLimitFromRow(row)),
		})
	}

	return response, nil
}
