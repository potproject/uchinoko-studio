package memory

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/potproject/uchinoko-studio/data"
	"github.com/potproject/uchinoko-studio/db"
	"github.com/potproject/uchinoko-studio/db/sqlcgen"
)

func setupMemoryTestDB(t *testing.T) {
	t.Helper()
	sqlitePath := filepath.Join(t.TempDir(), "database")
	if err := db.StartWithPath(sqlitePath); err != nil {
		t.Fatalf("db.StartWithPath() error = %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Fatalf("db.Close() error = %v", err)
		}
	})
}

func sampleMemoryCharacter() data.CharacterConfig {
	return data.CharacterConfig{
		General: data.CharacterConfigGeneral{
			ID:   "memory-char",
			Name: "Memory",
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
			SystemPrompt: "test",
		},
		Memory: data.CharacterConfigMemory{
			Enabled:                  true,
			MaxItemsInPrompt:         6,
			EnableRelationshipMemory: true,
			EnableSessionSummary:     true,
			EnableSemanticSearch:     false,
			EmbeddingModel:           "text-embedding-3-small",
			AllowSensitiveMemory:     false,
		},
	}
}

func TestSearchMemoryCandidatesUsesJapaneseTrigram(t *testing.T) {
	setupMemoryTestDB(t)

	character := sampleMemoryCharacter()
	if err := db.PutCharacterConfig(character.General.ID, character); err != nil {
		t.Fatalf("db.PutCharacterConfig() error = %v", err)
	}

	if _, err := CreateMemoryItem(data.MemoryItem{
		CharacterID:  character.General.ID,
		Scope:        data.MemoryScopeCharacter,
		Kind:         "persona_rule",
		Content:      "今日はいい天気だと感じたら明るく話す",
		KeywordsText: "天気 明るい",
		Pinned:       true,
		Confidence:   1,
		Salience:     1,
	}); err != nil {
		t.Fatalf("CreateMemoryItem(character) error = %v", err)
	}

	if _, err := CreateMemoryItem(data.MemoryItem{
		CharacterID:  character.General.ID,
		OwnerID:      "owner-1",
		Scope:        data.MemoryScopeRelationship,
		Kind:         "profile_fact",
		Content:      "ユーザーの名前は太郎",
		KeywordsText: "名前 太郎",
		Confidence:   0.9,
		Salience:     0.8,
	}); err != nil {
		t.Fatalf("CreateMemoryItem(relationship) error = %v", err)
	}

	candidates, err := SearchMemoryCandidates("owner-1", character.General.ID, "いい天", character.Memory)
	if err != nil {
		t.Fatalf("SearchMemoryCandidates() error = %v", err)
	}
	if len(candidates) == 0 {
		t.Fatal("SearchMemoryCandidates() len = 0, want >= 1")
	}
	if candidates[0].Item.Scope != data.MemoryScopeCharacter {
		t.Fatalf("SearchMemoryCandidates()[0].scope = %s, want %s", candidates[0].Item.Scope, data.MemoryScopeCharacter)
	}
}

func TestSearchMemoryCandidatesUsesJapaneseConversationFragments(t *testing.T) {
	setupMemoryTestDB(t)

	character := sampleMemoryCharacter()
	if err := db.PutCharacterConfig(character.General.ID, character); err != nil {
		t.Fatalf("db.PutCharacterConfig() error = %v", err)
	}

	if _, err := CreateMemoryItem(data.MemoryItem{
		CharacterID:  character.General.ID,
		OwnerID:      "owner-1",
		Scope:        data.MemoryScopeRelationship,
		Kind:         "preference",
		Content:      "ユーザーは辛いものが好き",
		KeywordsText: "辛いもの 好き",
		Confidence:   0.9,
		Salience:     0.8,
	}); err != nil {
		t.Fatalf("CreateMemoryItem(relationship) error = %v", err)
	}

	query := "最近どう？\n辛いもの好きって言ってたよね\n今夜は何を食べたい？"
	candidates, err := SearchMemoryCandidates("owner-1", character.General.ID, query, character.Memory)
	if err != nil {
		t.Fatalf("SearchMemoryCandidates() error = %v", err)
	}
	if len(candidates) == 0 {
		t.Fatal("SearchMemoryCandidates() len = 0, want >= 1")
	}
	if candidates[0].Item.Content != "ユーザーは辛いものが好き" {
		t.Fatalf("SearchMemoryCandidates()[0].content = %q, want relationship memory", candidates[0].Item.Content)
	}
}

func TestEnqueueCompactSessionDeduplicatesByUniqueKey(t *testing.T) {
	setupMemoryTestDB(t)

	if err := EnqueueCompactSession(CompactSessionPayload{
		OwnerID:     "owner-1",
		CharacterID: "character-1",
		SessionID:   "session-1",
		MaxHistory:  8,
	}); err != nil {
		t.Fatalf("EnqueueCompactSession(first) error = %v", err)
	}
	if err := EnqueueCompactSession(CompactSessionPayload{
		OwnerID:     "owner-1",
		CharacterID: "character-1",
		SessionID:   "session-1",
		MaxHistory:  8,
	}); err != nil {
		t.Fatalf("EnqueueCompactSession(second) error = %v", err)
	}

	var count int
	countValue, err := sqlcgen.New(db.SQLDB()).CountMemoryJobs(context.Background())
	if err != nil {
		t.Fatalf("count memory_jobs error = %v", err)
	}
	count = int(countValue)
	if count != 1 {
		t.Fatalf("memory_jobs count = %d, want 1", count)
	}
}
