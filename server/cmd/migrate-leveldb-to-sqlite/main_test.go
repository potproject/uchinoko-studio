package main

import (
	"encoding/json"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/potproject/uchinoko-studio/data"
	"github.com/potproject/uchinoko-studio/db"
	"github.com/syndtr/goleveldb/leveldb"
)

func sampleLegacyCharacter(id string, name string) data.CharacterConfig {
	return data.CharacterConfig{
		General: data.CharacterConfigGeneral{
			ID:   id,
			Name: name,
		},
		MultiVoice: false,
		Voice: []data.CharacterConfigVoice{
			{
				Name:      "Default",
				Type:      "voicevox",
				SpeakerID: "1",
				Image:     "default.png",
			},
		},
		Chat: data.CharacterConfigChat{
			Type:         "openai",
			Model:        "gpt-4o-mini",
			SystemPrompt: "legacy prompt",
		},
		Memory: data.CharacterConfigMemory{
			Enabled:                  false,
			MaxItemsInPrompt:         6,
			EnableRelationshipMemory: true,
			EnableSessionSummary:     true,
			EnableSemanticSearch:     true,
			EmbeddingModel:           "text-embedding-3-small",
			AllowSensitiveMemory:     false,
		},
	}
}

func putLegacyJSON(t *testing.T, legacyDB *leveldb.DB, key string, value any) {
	t.Helper()

	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("json.Marshal(%s) error = %v", key, err)
	}

	if err := legacyDB.Put([]byte(key), encoded, nil); err != nil {
		t.Fatalf("legacyDB.Put(%s) error = %v", key, err)
	}
}

func TestMigrateCopiesLevelDBData(t *testing.T) {
	tempDir := t.TempDir()
	levelDBPath := filepath.Join(tempDir, "legacy")
	sqlitePath := filepath.Join(tempDir, "sqlite.db")
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Fatalf("db.Close() error = %v", err)
		}
	})

	legacyDB, err := leveldb.OpenFile(levelDBPath, nil)
	if err != nil {
		t.Fatalf("leveldb.OpenFile() error = %v", err)
	}

	general := data.GeneralConfig{
		Background:            "blue",
		Language:              "ja-JP",
		SoundEffect:           true,
		CharacterOutputChange: false,
		EnableTTSOptimization: true,
	}
	general.Transcription.Type = "openai_speech_to_text"
	general.Transcription.Method = "auto"
	general.Transcription.AutoSetting.Threshold = 0.1
	general.Transcription.AutoSetting.SilentThreshold = 1.2
	general.Transcription.AutoSetting.AudioMinLength = 1.5

	envConfig := data.EnvConfig{
		OPENAI_API_KEY:         "openai-key",
		GEMINI_API_KEY:         "gemini-key",
		VOICEVOX_ENDPOINT:      "http://localhost:50021/",
		OPENAI_SPEECH_API_KEY:  "speech-key",
		VOSK_SERVER_ENDPOINT:   "localhost:2700",
		STYLEBERTVIT2_ENDPOINT: "http://localhost:8000/",
	}

	character := sampleLegacyCharacter("character-1", "Migrated")
	chatMessage := data.ChatMessage{
		Chat: []data.ChatCompletionMessage{
			{Role: data.ChatCompletionMessageRoleUser, Content: "hello"},
			{Role: data.ChatCompletionMessageRoleAssistant, Content: "world"},
		},
	}
	rateLimit := data.RateLimit{
		Day:    data.RateLimitType{LastUpdate: "20260329", Request: 2, Token: 20},
		Hour:   data.RateLimitType{LastUpdate: "2026032910", Request: 1, Token: 10},
		Minute: data.RateLimitType{LastUpdate: "202603291000", Request: 1, Token: 10},
	}

	putLegacyJSON(t, legacyDB, generalConfigKey, general)
	putLegacyJSON(t, legacyDB, envConfigKey, envConfig)
	putLegacyJSON(t, legacyDB, characterConfigKey, data.CharacterConfigList{
		Characters: []data.CharacterConfig{character},
	})
	putLegacyJSON(t, legacyDB, "session-1/character-1/chatmessage", chatMessage)
	putLegacyJSON(t, legacyDB, "rate_limit_session-1", rateLimit)

	if err := legacyDB.Close(); err != nil {
		t.Fatalf("legacyDB.Close() error = %v", err)
	}

	if err := migrate(levelDBPath, sqlitePath); err != nil {
		t.Fatalf("migrate() error = %v", err)
	}

	gotGeneral, err := db.GetGeneralConfig()
	if err != nil {
		t.Fatalf("db.GetGeneralConfig() error = %v", err)
	}
	if !reflect.DeepEqual(gotGeneral, general) {
		t.Fatalf("db.GetGeneralConfig() = %#v, want %#v", gotGeneral, general)
	}

	gotEnv, err := db.GetEnvConfig()
	if err != nil {
		t.Fatalf("db.GetEnvConfig() error = %v", err)
	}
	if !reflect.DeepEqual(gotEnv, envConfig) {
		t.Fatalf("db.GetEnvConfig() = %#v, want %#v", gotEnv, envConfig)
	}

	gotCharacter, err := db.GetCharacterConfig(character.General.ID)
	if err != nil {
		t.Fatalf("db.GetCharacterConfig() error = %v", err)
	}
	if !reflect.DeepEqual(gotCharacter, character) {
		t.Fatalf("db.GetCharacterConfig() = %#v, want %#v", gotCharacter, character)
	}

	gotChat, empty, err := db.GetChatMessage("session-1", "character-1")
	if err != nil {
		t.Fatalf("db.GetChatMessage() error = %v", err)
	}
	if empty {
		t.Fatal("db.GetChatMessage() empty = true, want false")
	}
	if !reflect.DeepEqual(gotChat, chatMessage) {
		t.Fatalf("db.GetChatMessage() = %#v, want %#v", gotChat, chatMessage)
	}

	all, err := db.ListAll()
	if err != nil {
		t.Fatalf("db.ListAll() error = %v", err)
	}
	if len(all) != 5 {
		t.Fatalf("db.ListAll() len = %d, want 5", len(all))
	}
}

func TestMigrateFailsWhenSQLiteAlreadyHasData(t *testing.T) {
	tempDir := t.TempDir()
	levelDBPath := filepath.Join(tempDir, "legacy")
	sqlitePath := filepath.Join(tempDir, "sqlite.db")
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Fatalf("db.Close() error = %v", err)
		}
	})

	legacyDB, err := leveldb.OpenFile(levelDBPath, nil)
	if err != nil {
		t.Fatalf("leveldb.OpenFile() error = %v", err)
	}
	putLegacyJSON(t, legacyDB, generalConfigKey, data.GeneralConfig{})
	if err := legacyDB.Close(); err != nil {
		t.Fatalf("legacyDB.Close() error = %v", err)
	}

	if err := db.StartWithPath(sqlitePath); err != nil {
		t.Fatalf("db.StartWithPath() error = %v", err)
	}
	if err := db.PutGeneralConfig(data.GeneralConfig{Background: "existing"}); err != nil {
		t.Fatalf("db.PutGeneralConfig() error = %v", err)
	}

	if err := migrate(levelDBPath, sqlitePath); err == nil {
		t.Fatal("migrate() error = nil, want non-nil")
	}
}

func TestMigrateSkipsLegacyChatKeyWithoutCharacterID(t *testing.T) {
	tempDir := t.TempDir()
	levelDBPath := filepath.Join(tempDir, "legacy")
	sqlitePath := filepath.Join(tempDir, "sqlite.db")
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Fatalf("db.Close() error = %v", err)
		}
	})

	legacyDB, err := leveldb.OpenFile(levelDBPath, nil)
	if err != nil {
		t.Fatalf("leveldb.OpenFile() error = %v", err)
	}

	character := sampleLegacyCharacter("character-1", "Default")
	putLegacyJSON(t, legacyDB, characterConfigKey, data.CharacterConfigList{
		Characters: []data.CharacterConfig{character},
	})
	putLegacyJSON(t, legacyDB, "session-1/chatmessage", data.ChatMessage{
		Chat: []data.ChatCompletionMessage{
			{Role: data.ChatCompletionMessageRoleUser, Content: "legacy"},
		},
	})

	if err := legacyDB.Close(); err != nil {
		t.Fatalf("legacyDB.Close() error = %v", err)
	}

	if err := migrate(levelDBPath, sqlitePath); err != nil {
		t.Fatalf("migrate() error = %v", err)
	}

	gotChat, empty, err := db.GetChatMessage("session-1", character.General.ID)
	if err != nil {
		t.Fatalf("db.GetChatMessage() error = %v", err)
	}
	if !empty {
		t.Fatalf("db.GetChatMessage() empty = %v, want true", empty)
	}
	if len(gotChat.Chat) != 0 {
		t.Fatalf("db.GetChatMessage() = %#v, want empty chat after skip", gotChat)
	}

	gotCharacter, err := db.GetCharacterConfig(character.General.ID)
	if err != nil {
		t.Fatalf("db.GetCharacterConfig() error = %v", err)
	}
	if !reflect.DeepEqual(gotCharacter, character) {
		t.Fatalf("db.GetCharacterConfig() = %#v, want %#v", gotCharacter, character)
	}
}
