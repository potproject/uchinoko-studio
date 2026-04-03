package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"github.com/potproject/uchinoko-studio/data"
	"github.com/potproject/uchinoko-studio/db"
	"github.com/potproject/uchinoko-studio/envgen"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

const (
	characterConfigKey = "character_config"
	generalConfigKey   = "general_config"
	envConfigKey       = "env_config"
	rateLimitPrefix    = "rate_limit_"
	chatMessageSuffix  = "/chatmessage"
)

func main() {
	if err := loadEnv(); err != nil {
		log.Fatal(err)
	}

	defaultLevelDBPath := envgen.Get().DB_FILE_PATH()
	defaultSQLitePath := resolveSQLitePath(defaultLevelDBPath)

	levelDBPath := flag.String("leveldb-path", defaultLevelDBPath, "existing LevelDB path")
	sqlitePath := flag.String("sqlite-path", defaultSQLitePath, "destination SQLite path")
	flag.Parse()

	if err := migrate(*levelDBPath, resolveSQLitePath(*sqlitePath)); err != nil {
		log.Fatal(err)
	}
}

func migrate(levelDBPath string, sqlitePath string) error {
	legacyDB, err := leveldb.OpenFile(levelDBPath, &opt.Options{ReadOnly: true})
	if err != nil {
		return fmt.Errorf("open leveldb: %w", err)
	}
	defer legacyDB.Close()

	if err := db.StartWithPath(sqlitePath); err != nil {
		return fmt.Errorf("start sqlite db: %w", err)
	}

	existing, err := db.ListAll()
	if err != nil {
		return fmt.Errorf("inspect sqlite database: %w", err)
	}
	if len(existing) != 0 {
		return fmt.Errorf("sqlite database already contains %d records", len(existing))
	}

	iter := legacyDB.NewIterator(nil, nil)
	defer iter.Release()

	for iter.Next() {
		key := string(iter.Key())
		value := iter.Value()

		if err := migrateEntry(key, value); err != nil {
			return fmt.Errorf("migrate %s: %w", key, err)
		}
	}

	if err := iter.Error(); err != nil {
		return fmt.Errorf("iterate leveldb: %w", err)
	}

	return nil
}

func migrateEntry(key string, value []byte) error {
	switch {
	case key == generalConfigKey:
		var config data.GeneralConfig
		if err := json.Unmarshal(value, &config); err != nil {
			return err
		}
		return db.PutGeneralConfig(config)
	case key == envConfigKey:
		var config data.EnvConfig
		if err := json.Unmarshal(value, &config); err != nil {
			return err
		}
		return db.PutEnvConfig(config)
	case key == characterConfigKey:
		var list data.CharacterConfigList
		if err := json.Unmarshal(value, &list); err != nil {
			return err
		}
		for _, character := range list.Characters {
			if err := db.PutCharacterConfig(character.General.ID, character); err != nil {
				return err
			}
		}
		return nil
	case strings.HasSuffix(key, chatMessageSuffix):
		sessionID, characterID, ok, err := parseChatMessageKey(key)
		if err != nil {
			return err
		}
		if !ok {
			log.Printf("skip legacy chat key without character id: %s", key)
			return nil
		}
		var message data.ChatMessage
		if err := json.Unmarshal(value, &message); err != nil {
			return err
		}
		return db.PutChatMessage(sessionID, characterID, message)
	case strings.HasPrefix(key, rateLimitPrefix):
		id := strings.TrimPrefix(key, rateLimitPrefix)
		var limit data.RateLimit
		if err := json.Unmarshal(value, &limit); err != nil {
			return err
		}
		return db.PutRateLimitSnapshot(id, limit)
	default:
		log.Printf("skip unknown key: %s", key)
		return nil
	}
}

func parseChatMessageKey(key string) (string, string, bool, error) {
	trimmed := strings.TrimSuffix(key, chatMessageSuffix)
	parts := strings.SplitN(trimmed, "/", 2)
	if len(parts) == 2 {
		return parts[0], parts[1], true, nil
	}

	if len(parts) == 1 && parts[0] != "" {
		return "", "", false, nil
	}

	return "", "", false, fmt.Errorf("invalid chat message key: %s", key)
}

func resolveSQLitePath(rawPath string) string {
	rawPath = strings.TrimSpace(rawPath)
	if rawPath == "" {
		rawPath = "database"
	}

	cleaned := filepath.Clean(rawPath)
	ext := strings.ToLower(filepath.Ext(cleaned))
	if ext == ".db" || ext == ".sqlite" {
		return cleaned
	}

	return filepath.Join(cleaned, "uchinoko.db")
}

func loadEnv() error {
	envFile := ".env"
	if _, err := os.Stat(envFile); err != nil {
		envFile = "env.txt"
	}
	if err := godotenv.Load(envFile); err != nil {
		log.Println("Not loading .env file")
	}
	return envgen.Load()
}
