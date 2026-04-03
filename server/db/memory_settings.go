package db

import (
	"context"
	"database/sql"

	"github.com/potproject/uchinoko-studio/data"
	"github.com/potproject/uchinoko-studio/db/sqlcgen"
)

func memoryConfigInit() data.CharacterConfigMemory {
	return data.CharacterConfigMemory{
		Enabled:                  false,
		MaxItemsInPrompt:         6,
		EnableRelationshipMemory: true,
		EnableSessionSummary:     true,
		EnableSemanticSearch:     true,
		EmbeddingModel:           "text-embedding-3-small",
		AllowSensitiveMemory:     false,
	}
}

func getCharacterMemoryConfig(characterID string) data.CharacterConfigMemory {
	if db == nil {
		return memoryConfigInit()
	}

	row, err := queries.GetCharacterMemorySetting(context.Background(), characterID)
	if err != nil {
		return memoryConfigInit()
	}

	return data.CharacterConfigMemory{
		Enabled:                  intToBool(row.Enabled),
		MaxItemsInPrompt:         row.MaxItemsInPrompt,
		EnableRelationshipMemory: intToBool(row.EnableRelationshipMemory),
		EnableSessionSummary:     intToBool(row.EnableSessionSummary),
		EnableSemanticSearch:     intToBool(row.EnableSemanticSearch),
		EmbeddingModel:           row.EmbeddingModel,
		AllowSensitiveMemory:     intToBool(row.AllowSensitiveMemory),
	}
}

func putCharacterMemoryConfigTx(ctx context.Context, tx *sql.Tx, characterID string, config data.CharacterConfigMemory) error {
	if config.MaxItemsInPrompt <= 0 {
		config.MaxItemsInPrompt = memoryConfigInit().MaxItemsInPrompt
	}
	if config.EmbeddingModel == "" {
		config.EmbeddingModel = memoryConfigInit().EmbeddingModel
	}
	return sqlcgen.New(tx).UpsertCharacterMemorySetting(ctx, sqlcgen.UpsertCharacterMemorySettingParams{
		CharacterID:              characterID,
		Enabled:                  boolToInt(config.Enabled),
		MaxItemsInPrompt:         config.MaxItemsInPrompt,
		EnableRelationshipMemory: boolToInt(config.EnableRelationshipMemory),
		EnableSessionSummary:     boolToInt(config.EnableSessionSummary),
		EnableSemanticSearch:     boolToInt(config.EnableSemanticSearch),
		EmbeddingModel:           config.EmbeddingModel,
		AllowSensitiveMemory:     boolToInt(config.AllowSensitiveMemory),
	})
}
