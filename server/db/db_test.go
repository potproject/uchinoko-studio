package db

import (
	"database/sql"
	"errors"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/potproject/uchinoko-studio/data"
)

func setupTestDB(t *testing.T) string {
	t.Helper()

	timeNow = time.Now

	sqlitePath := filepath.Join(t.TempDir(), "database")
	if err := StartWithPath(sqlitePath); err != nil {
		t.Fatalf("StartWithPath() error = %v", err)
	}

	t.Cleanup(func() {
		timeNow = time.Now
		if err := closeCurrentDB(); err != nil {
			t.Fatalf("closeCurrentDB() error = %v", err)
		}
	})

	return sqlitePath
}

func sampleCharacterConfig(id string, name string) data.CharacterConfig {
	return data.CharacterConfig{
		General: data.CharacterConfigGeneral{
			ID:   id,
			Name: name,
		},
		MultiVoice: true,
		Voice: []data.CharacterConfigVoice{
			{
				Name:                "Main",
				Type:                "voicevox",
				Identification:      "main",
				ModelID:             "model-1",
				ModelFile:           "voice.bin",
				SpeakerID:           "3",
				ReferenceAudioPath:  "refs/sample.wav",
				Image:               "main.png",
				BackgroundImagePath: "bg.png",
				Behavior: []data.CharacterConfigVoiceBehavior{
					{
						Identification: "happy",
						ImagePath:      "happy.png",
					},
				},
			},
		},
		Chat: data.CharacterConfigChat{
			Type:         "openai",
			Model:        "gpt-4o-mini",
			SystemPrompt: "Be helpful.",
			Temperature: data.TemperatureConfig{
				Enable: true,
				Value:  0.8,
			},
			MaxHistory: 8,
			Limit: data.CharacterConfigChatLimit{
				Day:    data.CharacterConfigChatLimitType{Request: 10, Token: 100},
				Hour:   data.CharacterConfigChatLimitType{Request: 5, Token: 50},
				Minute: data.CharacterConfigChatLimitType{Request: 2, Token: 20},
			},
		},
	}
}

func TestFreshStartDefaults(t *testing.T) {
	setupTestDB(t)

	general, err := GetGeneralConfig()
	if err != nil {
		t.Fatalf("GetGeneralConfig() error = %v", err)
	}
	if !reflect.DeepEqual(general, generalInitConfig()) {
		t.Fatalf("GetGeneralConfig() = %#v, want %#v", general, generalInitConfig())
	}

	envConfig, err := GetEnvConfig()
	if err != nil {
		t.Fatalf("GetEnvConfig() error = %v", err)
	}
	if !reflect.DeepEqual(envConfig, envInitConfig()) {
		t.Fatalf("GetEnvConfig() = %#v, want %#v", envConfig, envInitConfig())
	}

	characters, err := GetCharacterConfigList()
	if err != nil {
		t.Fatalf("GetCharacterConfigList() error = %v", err)
	}
	if len(characters.Characters) != 0 {
		t.Fatalf("GetCharacterConfigList() len = %d, want 0", len(characters.Characters))
	}

	chatMessage, empty, err := GetChatMessage("session", "character")
	if err != nil {
		t.Fatalf("GetChatMessage() error = %v", err)
	}
	if !empty {
		t.Fatalf("GetChatMessage() empty = %v, want true", empty)
	}
	if !reflect.DeepEqual(chatMessage, initChatMessage()) {
		t.Fatalf("GetChatMessage() = %#v, want %#v", chatMessage, initChatMessage())
	}

	allowed, err := RateLimitIsAllowed("session", data.CharacterConfigChatLimit{
		Day:    data.CharacterConfigChatLimitType{Request: 1, Token: 1},
		Hour:   data.CharacterConfigChatLimitType{Request: 1, Token: 1},
		Minute: data.CharacterConfigChatLimitType{Request: 1, Token: 1},
	})
	if err != nil {
		t.Fatalf("RateLimitIsAllowed() error = %v", err)
	}
	if !allowed {
		t.Fatal("RateLimitIsAllowed() = false, want true")
	}

	var gooseVersion int64
	if err := db.QueryRow("SELECT MAX(version_id) FROM goose_db_version").Scan(&gooseVersion); err != nil {
		t.Fatalf("query goose_db_version error = %v", err)
	}
	if gooseVersion != 1 {
		t.Fatalf("goose version = %d, want 1", gooseVersion)
	}
}

func TestStartWithExistingLegacyTablesSucceeds(t *testing.T) {
	sqlitePath := filepath.Join(t.TempDir(), "database")
	resolvedPath, err := resolveSQLitePath(sqlitePath)
	if err != nil {
		t.Fatalf("resolveSQLitePath() error = %v", err)
	}

	conn, err := openSQLite(resolvedPath)
	if err != nil {
		t.Fatalf("openSQLite() error = %v", err)
	}

	legacyStatements := []string{
		"CREATE TABLE IF NOT EXISTS general_config (id INTEGER PRIMARY KEY CHECK (id = 1), background TEXT NOT NULL, language TEXT NOT NULL, sound_effect INTEGER NOT NULL, character_output_change INTEGER NOT NULL, enable_tts_optimization INTEGER NOT NULL, transcription_type TEXT NOT NULL, transcription_method TEXT NOT NULL, transcription_auto_threshold REAL NOT NULL, transcription_auto_silent_threshold REAL NOT NULL, transcription_auto_audio_min_length REAL NOT NULL)",
		"CREATE TABLE IF NOT EXISTS env_config (id INTEGER PRIMARY KEY CHECK (id = 1), openai_speech_to_text_api_key TEXT NOT NULL, google_speech_to_text_api_key TEXT NOT NULL, vosk_server_endpoint TEXT NOT NULL, openai_api_key TEXT NOT NULL, anthropic_api_key TEXT NOT NULL, deepseek_api_key TEXT NOT NULL, gemini_api_key TEXT NOT NULL, openai_local_api_key TEXT NOT NULL, openai_local_api_endpoint TEXT NOT NULL, voicevox_endpoint TEXT NOT NULL, bertvits2_endpoint TEXT NOT NULL, irodori_tts_endpoint TEXT NOT NULL, nijivoice_api_key TEXT NOT NULL, stylebertvit2_endpoint TEXT NOT NULL, google_text_to_speech_api_key TEXT NOT NULL, openai_speech_api_key TEXT NOT NULL)",
		"CREATE TABLE IF NOT EXISTS characters (id TEXT PRIMARY KEY, name TEXT NOT NULL, multi_voice INTEGER NOT NULL, voice_json TEXT NOT NULL, chat_json TEXT NOT NULL)",
		"CREATE TABLE IF NOT EXISTS chat_sessions (session_id TEXT NOT NULL, character_id TEXT NOT NULL, messages_json TEXT NOT NULL, PRIMARY KEY (session_id, character_id))",
		"CREATE TABLE IF NOT EXISTS rate_limits (id TEXT PRIMARY KEY, day_last_update TEXT NOT NULL, day_request INTEGER NOT NULL, day_token INTEGER NOT NULL, hour_last_update TEXT NOT NULL, hour_request INTEGER NOT NULL, hour_token INTEGER NOT NULL, minute_last_update TEXT NOT NULL, minute_request INTEGER NOT NULL, minute_token INTEGER NOT NULL)",
	}
	for _, stmt := range legacyStatements {
		if _, err := conn.Exec(stmt); err != nil {
			t.Fatalf("seed legacy schema error = %v", err)
		}
	}
	if err := conn.Close(); err != nil {
		t.Fatalf("close legacy db error = %v", err)
	}

	if err := StartWithPath(sqlitePath); err != nil {
		t.Fatalf("StartWithPath(legacy) error = %v", err)
	}

	t.Cleanup(func() {
		if err := closeCurrentDB(); err != nil {
			t.Fatalf("closeCurrentDB() error = %v", err)
		}
	})

	var gooseVersion int64
	err = db.QueryRow("SELECT MAX(version_id) FROM goose_db_version").Scan(&gooseVersion)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("query goose_db_version error = %v", err)
	}
	if gooseVersion != 1 {
		t.Fatalf("goose version = %d, want 1", gooseVersion)
	}
}

func TestGeneralAndEnvRoundTrip(t *testing.T) {
	setupTestDB(t)

	general := generalInitConfig()
	general.Background = "sunset"
	general.Language = "en-US"
	general.SoundEffect = false
	general.EnableTTSOptimization = true
	general.Transcription.Type = "vosk"
	general.Transcription.AutoSetting.AudioMinLength = 2.5

	if err := PutGeneralConfig(general); err != nil {
		t.Fatalf("PutGeneralConfig() error = %v", err)
	}

	gotGeneral, err := GetGeneralConfig()
	if err != nil {
		t.Fatalf("GetGeneralConfig() error = %v", err)
	}
	if !reflect.DeepEqual(gotGeneral, general) {
		t.Fatalf("GetGeneralConfig() = %#v, want %#v", gotGeneral, general)
	}

	envConfig := envInitConfig()
	envConfig.OPENAI_API_KEY = "openai-key"
	envConfig.GEMINI_API_KEY = "gemini-key"
	envConfig.STYLEBERTVIT2_ENDPOINT = "http://localhost:8000/"

	if err := PutEnvConfig(envConfig); err != nil {
		t.Fatalf("PutEnvConfig() error = %v", err)
	}

	gotEnvConfig, err := GetEnvConfig()
	if err != nil {
		t.Fatalf("GetEnvConfig() error = %v", err)
	}
	if !reflect.DeepEqual(gotEnvConfig, envConfig) {
		t.Fatalf("GetEnvConfig() = %#v, want %#v", gotEnvConfig, envConfig)
	}
}

func TestCharacterCRUD(t *testing.T) {
	setupTestDB(t)

	first := sampleCharacterConfig("char-b", "Beta")
	second := sampleCharacterConfig("char-a", "Alpha")

	if err := PutCharacterConfig(first.General.ID, first); err != nil {
		t.Fatalf("PutCharacterConfig(first) error = %v", err)
	}
	if err := PutCharacterConfig(second.General.ID, second); err != nil {
		t.Fatalf("PutCharacterConfig(second) error = %v", err)
	}

	gotSecond, err := GetCharacterConfig(second.General.ID)
	if err != nil {
		t.Fatalf("GetCharacterConfig() error = %v", err)
	}
	if !reflect.DeepEqual(gotSecond, second) {
		t.Fatalf("GetCharacterConfig() = %#v, want %#v", gotSecond, second)
	}

	list, err := GetCharacterConfigList()
	if err != nil {
		t.Fatalf("GetCharacterConfigList() error = %v", err)
	}
	if len(list.Characters) != 2 {
		t.Fatalf("GetCharacterConfigList() len = %d, want 2", len(list.Characters))
	}
	if list.Characters[0].General.ID != second.General.ID || list.Characters[1].General.ID != first.General.ID {
		t.Fatalf("GetCharacterConfigList() order = %#v", list.Characters)
	}

	second.General.Name = "Alpha Updated"
	second.Chat.Model = "gpt-4.1-mini"
	if err := PutCharacterConfig(second.General.ID, second); err != nil {
		t.Fatalf("PutCharacterConfig(update) error = %v", err)
	}

	updated, err := GetCharacterConfig(second.General.ID)
	if err != nil {
		t.Fatalf("GetCharacterConfig(update) error = %v", err)
	}
	if !reflect.DeepEqual(updated, second) {
		t.Fatalf("GetCharacterConfig(update) = %#v, want %#v", updated, second)
	}

	if err := DeleteCharacterConfig(first.General.ID); err != nil {
		t.Fatalf("DeleteCharacterConfig() error = %v", err)
	}

	list, err = GetCharacterConfigList()
	if err != nil {
		t.Fatalf("GetCharacterConfigList(after delete) error = %v", err)
	}
	if len(list.Characters) != 1 || list.Characters[0].General.ID != second.General.ID {
		t.Fatalf("GetCharacterConfigList(after delete) = %#v", list.Characters)
	}
}

func TestChatMessageRoundTrip(t *testing.T) {
	setupTestDB(t)

	message := data.ChatMessage{
		Chat: []data.ChatCompletionMessage{
			{
				Role:    data.ChatCompletionMessageRoleUser,
				Content: "hello",
				Image: &data.Image{
					Extension: "png",
					Data:      []byte{0x01, 0x02, 0x03},
				},
			},
			{
				Role:    data.ChatCompletionMessageRoleAssistant,
				Content: "hi",
			},
		},
	}

	if err := PutChatMessage("session-1", "character-1", message); err != nil {
		t.Fatalf("PutChatMessage() error = %v", err)
	}

	got, empty, err := GetChatMessage("session-1", "character-1")
	if err != nil {
		t.Fatalf("GetChatMessage() error = %v", err)
	}
	if empty {
		t.Fatal("GetChatMessage() empty = true, want false")
	}
	if !reflect.DeepEqual(got, message) {
		t.Fatalf("GetChatMessage() = %#v, want %#v", got, message)
	}

	if err := DeleteChatMessage("session-1", "character-1"); err != nil {
		t.Fatalf("DeleteChatMessage() error = %v", err)
	}

	_, empty, err = GetChatMessage("session-1", "character-1")
	if err != nil {
		t.Fatalf("GetChatMessage(after delete) error = %v", err)
	}
	if !empty {
		t.Fatal("GetChatMessage(after delete) empty = false, want true")
	}
}

func TestRateLimitResetAndEnforcement(t *testing.T) {
	setupTestDB(t)

	base := time.Date(2026, 3, 29, 10, 0, 0, 0, time.UTC)
	timeNow = func() time.Time { return base }

	limit := data.CharacterConfigChatLimit{
		Day:    data.CharacterConfigChatLimitType{Request: 1, Token: 5},
		Hour:   data.CharacterConfigChatLimitType{Request: 1, Token: 5},
		Minute: data.CharacterConfigChatLimitType{Request: 1, Token: 5},
	}

	if err := AddRateLimit("user-1", 1, 1); err != nil {
		t.Fatalf("AddRateLimit() error = %v", err)
	}

	allowed, err := RateLimitIsAllowed("user-1", limit)
	if err != nil {
		t.Fatalf("RateLimitIsAllowed() error = %v", err)
	}
	if !allowed {
		t.Fatal("RateLimitIsAllowed() = false, want true")
	}

	timeNow = func() time.Time { return base.Add(24 * time.Hour) }

	allowed, err = RateLimitIsAllowed("user-1", limit)
	if err != nil {
		t.Fatalf("RateLimitIsAllowed(after reset) error = %v", err)
	}
	if !allowed {
		t.Fatal("RateLimitIsAllowed(after reset) = false, want true")
	}

	if err := AddRateLimit("user-1", 1, 1); err != nil {
		t.Fatalf("AddRateLimit(after reset) error = %v", err)
	}

	allowed, err = RateLimitIsAllowed("user-1", limit)
	if err != nil {
		t.Fatalf("RateLimitIsAllowed(second window) error = %v", err)
	}
	if !allowed {
		t.Fatal("RateLimitIsAllowed(second window) = false, want true")
	}

	if err := AddRateLimit("user-1", 1, 1); err != nil {
		t.Fatalf("AddRateLimit(exceed) error = %v", err)
	}

	allowed, err = RateLimitIsAllowed("user-1", limit)
	if err != nil {
		t.Fatalf("RateLimitIsAllowed(exceed) error = %v", err)
	}
	if allowed {
		t.Fatal("RateLimitIsAllowed(exceed) = true, want false")
	}
}

func TestListAllPseudoKeys(t *testing.T) {
	setupTestDB(t)

	if err := PutGeneralConfig(generalInitConfig()); err != nil {
		t.Fatalf("PutGeneralConfig() error = %v", err)
	}
	if err := PutEnvConfig(envInitConfig()); err != nil {
		t.Fatalf("PutEnvConfig() error = %v", err)
	}

	character := sampleCharacterConfig("character-1", "Sample")
	if err := PutCharacterConfig(character.General.ID, character); err != nil {
		t.Fatalf("PutCharacterConfig() error = %v", err)
	}

	if err := PutChatMessage("session-1", character.General.ID, data.ChatMessage{
		Chat: []data.ChatCompletionMessage{{Role: data.ChatCompletionMessageRoleUser, Content: "hi"}},
	}); err != nil {
		t.Fatalf("PutChatMessage() error = %v", err)
	}

	if err := PutRateLimitSnapshot("session-1", data.RateLimit{
		Day:    data.RateLimitType{LastUpdate: "20260329", Request: 1, Token: 2},
		Hour:   data.RateLimitType{LastUpdate: "2026032910", Request: 1, Token: 2},
		Minute: data.RateLimitType{LastUpdate: "202603291000", Request: 1, Token: 2},
	}); err != nil {
		t.Fatalf("PutRateLimitSnapshot() error = %v", err)
	}

	all, err := ListAll()
	if err != nil {
		t.Fatalf("ListAll() error = %v", err)
	}

	keys := map[string]bool{}
	for _, entry := range all {
		keys[entry.Key] = true
	}

	expectedKeys := []string{
		"general_config",
		"env_config",
		"character_config/character-1",
		"chatmessage/session-1/character-1",
		"rate_limit/session-1",
	}
	for _, key := range expectedKeys {
		if !keys[key] {
			t.Fatalf("ListAll() missing key %q in %#v", key, keys)
		}
	}
}
