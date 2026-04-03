package db

import (
	"context"
	"errors"
)

type MemoryLexicalCandidateRow struct {
	ID                int64
	PublicID          string
	CharacterID       string
	OwnerID           string
	Scope             string
	Kind              string
	Content           string
	KeywordsText      string
	Pinned            int64
	Confidence        float64
	Salience          float64
	Source            string
	UpdatedAt         string
	NormalizedContent string
	EmbeddingJSON     string
	LexicalRaw        float64
}

// sqlc could not interpret SQLite FTS MATCH syntax together with
// bm25(memory_items_fts), so this FTS query stays as a hand-written query in
// the db package and is kept out of the memory package.
func SearchMemoryLexicalCandidates(ctx context.Context, ownerID string, characterID string, matchQuery string, includeRelationship bool, limit int64) ([]MemoryLexicalCandidateRow, error) {
	if db == nil {
		return nil, errors.New("database is not started")
	}

	baseWhere := `
mi.character_id = ?
AND mi.superseded_by IS NULL
AND (
    (mi.scope = 'character' AND mi.owner_id = '')
`
	args := []any{characterID}
	if includeRelationship {
		baseWhere += ` OR (mi.scope = 'relationship' AND mi.owner_id = ?)`
		args = append(args, ownerID)
	}
	baseWhere += `)`

	sqlText := `
SELECT
    mi.id,
    mi.public_id,
    mi.character_id,
    mi.owner_id,
    mi.scope,
    mi.kind,
    mi.content,
    mi.keywords_text,
    mi.pinned,
    mi.confidence,
    mi.salience,
    mi.source,
    mi.updated_at,
    mi.normalized_content,
    mi.embedding_json,
    bm25(memory_items_fts) AS lexical_raw
FROM memory_items_fts
JOIN memory_items mi ON mi.id = memory_items_fts.rowid
WHERE ` + baseWhere + `
  AND memory_items_fts MATCH ?
ORDER BY lexical_raw
LIMIT ?`
	args = append(args, matchQuery, limit)

	rows, err := db.QueryContext(ctx, sqlText, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]MemoryLexicalCandidateRow, 0)
	for rows.Next() {
		var row MemoryLexicalCandidateRow
		err := rows.Scan(
			&row.ID,
			&row.PublicID,
			&row.CharacterID,
			&row.OwnerID,
			&row.Scope,
			&row.Kind,
			&row.Content,
			&row.KeywordsText,
			&row.Pinned,
			&row.Confidence,
			&row.Salience,
			&row.Source,
			&row.UpdatedAt,
			&row.NormalizedContent,
			&row.EmbeddingJSON,
			&row.LexicalRaw,
		)
		if err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
