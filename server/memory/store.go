package memory

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	"github.com/potproject/uchinoko-studio/data"
	"github.com/potproject/uchinoko-studio/db"
	"github.com/potproject/uchinoko-studio/db/sqlcgen"
	"golang.org/x/text/unicode/norm"
)

const (
	jobTypeExtractTurn          = "extract_turn"
	jobTypeCompactSession       = "compact_session"
	jobTypeReindexMemoryItem    = "reindex_memory_item"
	jobTypeRebuildCharacter     = "rebuild_character_memory"
	jobStatusQueued             = "queued"
	jobStatusRunning            = "running"
	jobStatusFailed             = "failed"
	maxMemorySearchCandidates   = 12
	defaultMemoryItemsInPrompt  = 6
	defaultWorkerRetryBaseDelay = time.Minute
)

var timeNow = time.Now

type memoryItemRecord struct {
	rowID int64
	data.MemoryItem
	normalized     string
	embedding      []float64
	supersededByID sql.NullInt64
}

type SearchCandidate struct {
	Item           data.MemoryItem
	RowID          int64
	Normalized     string
	Embedding      []float64
	LexicalScore   float64
	SemanticScore  float64
	RecencyScore   float64
	CombinedScore  float64
	PairwiseVector []float64
}

type Job struct {
	ID         int64
	Type       string
	UniqueKey  string
	PayloadRaw string
	Attempts   int64
}

type ExtractTurnPayload struct {
	OwnerID          string `json:"ownerId"`
	CharacterID      string `json:"characterId"`
	SessionID        string `json:"sessionId"`
	UserContent      string `json:"userContent"`
	AssistantContent string `json:"assistantContent"`
	UserIndex        int64  `json:"userIndex"`
	AssistantIndex   int64  `json:"assistantIndex"`
}

type CompactSessionPayload struct {
	OwnerID     string `json:"ownerId"`
	CharacterID string `json:"characterId"`
	SessionID   string `json:"sessionId"`
	MaxHistory  int64  `json:"maxHistory"`
}

type ReindexMemoryItemPayload struct {
	MemoryID string `json:"memoryId"`
}

type RebuildCharacterMemoryPayload struct {
	OwnerID     string `json:"ownerId"`
	CharacterID string `json:"characterId"`
}

func NormalizeText(text string) string {
	text = norm.NFKC.String(text)
	text = strings.Map(func(r rune) rune {
		if r <= unicode.MaxASCII {
			return unicode.ToLower(r)
		}
		return r
	}, text)
	text = strings.Join(strings.Fields(text), " ")
	return strings.TrimSpace(text)
}

func nowString() string {
	return timeNow().UTC().Format(time.RFC3339)
}

func withTx(ctx context.Context, fn func(*sql.Tx) error) error {
	conn := db.SQLDB()
	if conn == nil {
		return errors.New("database is not started")
	}
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

func memoryDefaults(config data.CharacterConfigMemory) data.CharacterConfigMemory {
	if config.MaxItemsInPrompt <= 0 {
		config.MaxItemsInPrompt = defaultMemoryItemsInPrompt
	}
	if config.EmbeddingModel == "" {
		config.EmbeddingModel = "text-embedding-3-small"
	}
	return config
}

func recordFromSQLC(item sqlcgen.MemoryItem) memoryItemRecord {
	rec := memoryItemRecord{
		rowID: item.ID,
		MemoryItem: data.MemoryItem{
			ID:           item.PublicID,
			CharacterID:  item.CharacterID,
			OwnerID:      item.OwnerID,
			Scope:        data.MemoryScope(item.Scope),
			Kind:         item.Kind,
			Content:      item.Content,
			KeywordsText: item.KeywordsText,
			Pinned:       item.Pinned != 0,
			Confidence:   item.Confidence,
			Salience:     item.Salience,
			Source:       item.Source,
			UpdatedAt:    item.UpdatedAt,
		},
		normalized:     item.NormalizedContent,
		supersededByID: item.SupersededBy,
	}
	if item.EmbeddingJson != "" {
		_ = json.Unmarshal([]byte(item.EmbeddingJson), &rec.embedding)
	}
	return rec
}

func candidateFromFallbackRow(row sqlcgen.SearchMemoryFallbackWithRelationshipRow) SearchCandidate {
	rec := recordFromSQLCFallback(row.ID, row.PublicID, row.CharacterID, row.OwnerID, row.Scope, row.Kind, row.Content, row.NormalizedContent, row.KeywordsText, row.Pinned, row.Confidence, row.Salience, row.Source, row.SupersededBy, row.EmbeddingJson, row.UpdatedAt)
	return SearchCandidate{
		Item:         rec.MemoryItem,
		RowID:        rec.rowID,
		Normalized:   rec.normalized,
		Embedding:    rec.embedding,
		LexicalScore: float64(row.LexicalRaw),
	}
}

func candidateFromFallbackCharacterRow(row sqlcgen.SearchMemoryFallbackCharacterOnlyRow) SearchCandidate {
	rec := recordFromSQLCFallback(row.ID, row.PublicID, row.CharacterID, row.OwnerID, row.Scope, row.Kind, row.Content, row.NormalizedContent, row.KeywordsText, row.Pinned, row.Confidence, row.Salience, row.Source, row.SupersededBy, row.EmbeddingJson, row.UpdatedAt)
	return SearchCandidate{
		Item:         rec.MemoryItem,
		RowID:        rec.rowID,
		Normalized:   rec.normalized,
		Embedding:    rec.embedding,
		LexicalScore: float64(row.LexicalRaw),
	}
}

func candidateFromDBSearchRow(row db.MemoryLexicalCandidateRow) SearchCandidate {
	rec := recordFromSQLCFallback(row.ID, row.PublicID, row.CharacterID, row.OwnerID, row.Scope, row.Kind, row.Content, row.NormalizedContent, row.KeywordsText, row.Pinned, row.Confidence, row.Salience, row.Source, sql.NullInt64{}, row.EmbeddingJSON, row.UpdatedAt)
	return SearchCandidate{
		Item:         rec.MemoryItem,
		RowID:        rec.rowID,
		Normalized:   rec.normalized,
		Embedding:    rec.embedding,
		LexicalScore: row.LexicalRaw,
	}
}

func recordFromSQLCFallback(rowID int64, publicID string, characterID string, ownerID string, scope string, kind string, content string, normalized string, keywords string, pinned int64, confidence float64, salience float64, source string, supersededBy sql.NullInt64, embeddingJSON string, updatedAt string) memoryItemRecord {
	rec := memoryItemRecord{
		rowID: rowID,
		MemoryItem: data.MemoryItem{
			ID:           publicID,
			CharacterID:  characterID,
			OwnerID:      ownerID,
			Scope:        data.MemoryScope(scope),
			Kind:         kind,
			Content:      content,
			KeywordsText: keywords,
			Pinned:       pinned != 0,
			Confidence:   confidence,
			Salience:     salience,
			Source:       source,
			UpdatedAt:    updatedAt,
		},
		normalized:     normalized,
		supersededByID: supersededBy,
	}
	if embeddingJSON != "" {
		_ = json.Unmarshal([]byte(embeddingJSON), &rec.embedding)
	}
	return rec
}

func ListMemoryItems(ownerID string, characterID string) (data.MemoryItemList, error) {
	if db.SQLDB() == nil {
		return data.MemoryItemList{}, errors.New("database is not started")
	}
	rows, err := sqlcgen.New(db.SQLDB()).ListMemoryItemsForOwnerAndCharacter(context.Background(), sqlcgen.ListMemoryItemsForOwnerAndCharacterParams{
		CharacterID: characterID,
		OwnerID:     ownerID,
	})
	if err != nil {
		return data.MemoryItemList{}, err
	}
	items := make([]data.MemoryItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, recordFromSQLC(row).MemoryItem)
	}
	return data.MemoryItemList{Items: items}, nil
}

func CreateMemoryItem(item data.MemoryItem) (data.MemoryItem, error) {
	item.Scope = data.MemoryScope(strings.TrimSpace(string(item.Scope)))
	if item.Scope == "" {
		item.Scope = data.MemoryScopeRelationship
	}
	if item.Scope == data.MemoryScopeCharacter {
		item.OwnerID = ""
		item.Pinned = true
	}
	if item.ID == "" {
		item.ID = uuid.New().String()
	}
	if item.Source == "" {
		item.Source = "manual"
	}
	if item.Kind == "" {
		item.Kind = "relationship_fact"
	}
	now := nowString()
	item.UpdatedAt = now
	normalized := NormalizeText(item.Content)
	keywords := NormalizeText(item.KeywordsText)
	var rowID int64
	err := withTx(context.Background(), func(tx *sql.Tx) error {
		res, err := sqlcgen.New(tx).InsertMemoryItem(context.Background(), sqlcgen.InsertMemoryItemParams{
			PublicID:          item.ID,
			CharacterID:       item.CharacterID,
			OwnerID:           item.OwnerID,
			Scope:             string(item.Scope),
			Kind:              item.Kind,
			Content:           item.Content,
			NormalizedContent: normalized,
			KeywordsText:      keywords,
			Pinned:            boolToInt(item.Pinned),
			Confidence:        item.Confidence,
			Salience:          item.Salience,
			Source:            item.Source,
			SupersededBy:      sql.NullInt64{},
			EmbeddingJson:     "",
			CreatedAt:         now,
			UpdatedAt:         now,
			LastAccessedAt:    now,
		})
		if err != nil {
			return err
		}
		rowID, err = res.LastInsertId()
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return data.MemoryItem{}, err
	}
	_ = EnqueueReindexMemoryItem(item.ID)
	_ = rowID
	return item, nil
}

func UpdateMemoryItem(itemID string, update data.MemoryItem) (data.MemoryItem, error) {
	if db.SQLDB() == nil {
		return data.MemoryItem{}, errors.New("database is not started")
	}
	current, rowID, err := getMemoryItemByPublicID(itemID)
	if err != nil {
		return data.MemoryItem{}, err
	}
	if strings.TrimSpace(update.Content) != "" {
		current.Content = update.Content
	}
	if strings.TrimSpace(update.KeywordsText) != "" || update.KeywordsText == "" {
		current.KeywordsText = update.KeywordsText
	}
	if strings.TrimSpace(update.Kind) != "" {
		current.Kind = update.Kind
	}
	current.Pinned = update.Pinned
	current.Confidence = update.Confidence
	current.Salience = update.Salience
	current.UpdatedAt = nowString()
	err = sqlcgen.New(db.SQLDB()).UpdateMemoryItemEditable(context.Background(), sqlcgen.UpdateMemoryItemEditableParams{
		Kind:              current.Kind,
		Content:           current.Content,
		NormalizedContent: NormalizeText(current.Content),
		KeywordsText:      NormalizeText(current.KeywordsText),
		Pinned:            boolToInt(current.Pinned),
		Confidence:        current.Confidence,
		Salience:          current.Salience,
		UpdatedAt:         current.UpdatedAt,
		LastAccessedAt:    current.UpdatedAt,
		ID:                rowID,
	})
	if err != nil {
		return data.MemoryItem{}, err
	}
	_ = EnqueueReindexMemoryItem(itemID)
	return current, nil
}

func DeleteMemoryItem(itemID string) error {
	if db.SQLDB() == nil {
		return errors.New("database is not started")
	}
	return sqlcgen.New(db.SQLDB()).DeleteMemoryItemByPublicID(context.Background(), itemID)
}

func GetSessionSummary(ownerID string, characterID string, sessionID string) (data.SessionSummary, error) {
	if db.SQLDB() == nil {
		return data.SessionSummary{}, errors.New("database is not started")
	}
	row, err := sqlcgen.New(db.SQLDB()).GetSessionSummary(context.Background(), sqlcgen.GetSessionSummaryParams{
		OwnerID:     ownerID,
		CharacterID: characterID,
		SessionID:   sessionID,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return data.SessionSummary{
			OwnerID:     ownerID,
			CharacterID: characterID,
			SessionID:   sessionID,
			Summary:     "",
			UpdatedAt:   "",
		}, nil
	}
	if err != nil {
		return data.SessionSummary{}, err
	}
	return data.SessionSummary{
		OwnerID:     row.OwnerID,
		CharacterID: row.CharacterID,
		SessionID:   row.SessionID,
		Summary:     row.Summary,
		UpdatedAt:   row.UpdatedAt,
	}, nil
}

func upsertSessionSummaryTx(tx *sql.Tx, summary data.SessionSummary) error {
	if summary.UpdatedAt == "" {
		summary.UpdatedAt = nowString()
	}
	return sqlcgen.New(tx).UpsertSessionSummary(context.Background(), sqlcgen.UpsertSessionSummaryParams{
		OwnerID:     summary.OwnerID,
		CharacterID: summary.CharacterID,
		SessionID:   summary.SessionID,
		Summary:     summary.Summary,
		UpdatedAt:   summary.UpdatedAt,
	})
}

func getMemoryItemByPublicID(itemID string) (data.MemoryItem, int64, error) {
	if db.SQLDB() == nil {
		return data.MemoryItem{}, 0, errors.New("database is not started")
	}
	row, err := sqlcgen.New(db.SQLDB()).GetMemoryItemByPublicID(context.Background(), itemID)
	if err != nil {
		return data.MemoryItem{}, 0, err
	}
	rec := recordFromSQLC(row)
	return rec.MemoryItem, rec.rowID, nil
}

func SearchMemoryCandidates(ownerID string, characterID string, query string, cfg data.CharacterConfigMemory) ([]SearchCandidate, error) {
	conn := db.SQLDB()
	if conn == nil {
		return nil, errors.New("database is not started")
	}
	cfg = memoryDefaults(cfg)
	query = NormalizeText(query)
	lexical, err := searchLexicalCandidates(conn, ownerID, characterID, query, cfg.EnableRelationshipMemory)
	if err != nil {
		return nil, err
	}
	pinned, err := listPinnedCandidates(conn, ownerID, characterID, cfg.EnableRelationshipMemory)
	if err != nil {
		return nil, err
	}
	merged := mergeCandidates(lexical, pinned)
	if len(merged) == 0 {
		return nil, nil
	}

	semanticEnabled := cfg.EnableSemanticSearch
	queryEmbedding := []float64(nil)
	if semanticEnabled {
		queryEmbedding, _ = embedText(query, cfg)
		if len(queryEmbedding) == 0 {
			semanticEnabled = false
		}
	}

	lexicalScores := normalizeLexicalScores(merged)
	for i := range merged {
		merged[i].LexicalScore = lexicalScores[i]
		merged[i].RecencyScore = recencyScore(merged[i].Item.UpdatedAt, merged[i].Item.Pinned)
		if semanticEnabled && len(merged[i].Embedding) > 0 {
			merged[i].SemanticScore = cosineSimilarity(queryEmbedding, merged[i].Embedding)
		}
		if semanticEnabled {
			merged[i].CombinedScore = 0.55*merged[i].LexicalScore + 0.35*merged[i].SemanticScore + 0.10*merged[i].RecencyScore
		} else {
			merged[i].CombinedScore = 0.85*merged[i].LexicalScore + 0.15*merged[i].RecencyScore
		}
	}

	sort.SliceStable(merged, func(i, j int) bool {
		if merged[i].Item.Pinned != merged[j].Item.Pinned {
			return merged[i].Item.Pinned
		}
		return merged[i].CombinedScore > merged[j].CombinedScore
	})

	if int(cfg.MaxItemsInPrompt) <= 0 {
		cfg.MaxItemsInPrompt = defaultMemoryItemsInPrompt
	}
	return applyMMR(merged, int(cfg.MaxItemsInPrompt)), nil
}

func searchLexicalCandidates(conn *sql.DB, ownerID string, characterID string, query string, includeRelationship bool) ([]SearchCandidate, error) {
	useFTS := utf8RuneCount(query) >= 3
	if query != "" && useFTS {
		matchQuery := buildFTSMatchQuery(query)
		rows, err := db.SearchMemoryLexicalCandidates(context.Background(), ownerID, characterID, matchQuery, includeRelationship, maxMemorySearchCandidates)
		if err != nil {
			return nil, err
		}
		out := make([]SearchCandidate, 0, len(rows))
		for _, row := range rows {
			out = append(out, candidateFromDBSearchRow(row))
		}
		return out, nil
	}
	if includeRelationship {
		rows, err := sqlcgen.New(conn).SearchMemoryFallbackWithRelationship(context.Background(), sqlcgen.SearchMemoryFallbackWithRelationshipParams{
			CharacterID: characterID,
			OwnerID:     ownerID,
			Query:       sql.NullString{String: query, Valid: true},
			LimitRows:   maxMemorySearchCandidates,
		})
		if err != nil {
			return nil, err
		}
		out := make([]SearchCandidate, 0, len(rows))
		for _, row := range rows {
			out = append(out, candidateFromFallbackRow(row))
		}
		return out, nil
	}
	rows, err := sqlcgen.New(conn).SearchMemoryFallbackCharacterOnly(context.Background(), sqlcgen.SearchMemoryFallbackCharacterOnlyParams{
		CharacterID: characterID,
		Query:       sql.NullString{String: query, Valid: true},
		LimitRows:   maxMemorySearchCandidates,
	})
	if err != nil {
		return nil, err
	}
	out := make([]SearchCandidate, 0, len(rows))
	for _, row := range rows {
		out = append(out, candidateFromFallbackCharacterRow(row))
	}
	return out, nil
}

func listPinnedCandidates(conn *sql.DB, ownerID string, characterID string, includeRelationship bool) ([]SearchCandidate, error) {
	if includeRelationship {
		rows, err := sqlcgen.New(conn).ListPinnedMemoryCandidatesWithRelationship(context.Background(), sqlcgen.ListPinnedMemoryCandidatesWithRelationshipParams{
			CharacterID: characterID,
			OwnerID:     ownerID,
			LimitRows:   4,
		})
		if err != nil {
			return nil, err
		}
		out := make([]SearchCandidate, 0, len(rows))
		for _, row := range rows {
			rec := recordFromSQLC(row)
			out = append(out, SearchCandidate{Item: rec.MemoryItem, RowID: rec.rowID, Normalized: rec.normalized, Embedding: rec.embedding})
		}
		return out, nil
	}
	rows, err := sqlcgen.New(conn).ListPinnedMemoryCandidatesCharacterOnly(context.Background(), sqlcgen.ListPinnedMemoryCandidatesCharacterOnlyParams{
		CharacterID: characterID,
		LimitRows:   4,
	})
	if err != nil {
		return nil, err
	}
	out := make([]SearchCandidate, 0, len(rows))
	for _, row := range rows {
		rec := recordFromSQLC(row)
		out = append(out, SearchCandidate{Item: rec.MemoryItem, RowID: rec.rowID, Normalized: rec.normalized, Embedding: rec.embedding})
	}
	return out, nil
}

func mergeCandidates(primary []SearchCandidate, extra []SearchCandidate) []SearchCandidate {
	seen := map[string]bool{}
	out := make([]SearchCandidate, 0, len(primary)+len(extra))
	for _, candidate := range append(primary, extra...) {
		if seen[candidate.Item.ID] {
			continue
		}
		seen[candidate.Item.ID] = true
		out = append(out, candidate)
	}
	return out
}

func normalizeLexicalScores(candidates []SearchCandidate) []float64 {
	if len(candidates) == 0 {
		return nil
	}
	minScore := candidates[0].LexicalScore
	maxScore := candidates[0].LexicalScore
	for _, candidate := range candidates[1:] {
		if candidate.LexicalScore < minScore {
			minScore = candidate.LexicalScore
		}
		if candidate.LexicalScore > maxScore {
			maxScore = candidate.LexicalScore
		}
	}
	out := make([]float64, len(candidates))
	if maxScore == minScore {
		for i := range out {
			out[i] = 1
		}
		return out
	}
	for i, candidate := range candidates {
		out[i] = 1 - ((candidate.LexicalScore - minScore) / (maxScore - minScore))
	}
	return out
}

func recencyScore(updatedAt string, pinned bool) float64 {
	if pinned {
		return 1
	}
	parsed, err := time.Parse(time.RFC3339, updatedAt)
	if err != nil {
		return 0.5
	}
	ageDays := timeNow().Sub(parsed).Hours() / 24
	if ageDays <= 0 {
		return 1
	}
	return 1 / (1 + (ageDays / 30))
}

func trigramSet(s string) map[string]struct{} {
	runes := []rune(s)
	if len(runes) < 3 {
		return map[string]struct{}{s: {}}
	}
	out := make(map[string]struct{}, len(runes)-2)
	for i := 0; i <= len(runes)-3; i++ {
		out[string(runes[i:i+3])] = struct{}{}
	}
	return out
}

func jaccardSimilarity(a string, b string) float64 {
	if a == "" || b == "" {
		return 0
	}
	sa := trigramSet(a)
	sb := trigramSet(b)
	intersection := 0
	for key := range sa {
		if _, ok := sb[key]; ok {
			intersection++
		}
	}
	union := len(sa) + len(sb) - intersection
	if union == 0 {
		return 0
	}
	return float64(intersection) / float64(union)
}

func cosineSimilarity(a []float64, b []float64) float64 {
	if len(a) == 0 || len(a) != len(b) {
		return 0
	}
	var dot float64
	var normA float64
	var normB float64
	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	score := dot / (math.Sqrt(normA) * math.Sqrt(normB))
	return (score + 1) / 2
}

func applyMMR(candidates []SearchCandidate, limit int) []SearchCandidate {
	if len(candidates) <= limit {
		return candidates
	}
	selected := make([]SearchCandidate, 0, limit)
	remaining := append([]SearchCandidate(nil), candidates...)
	for len(remaining) > 0 && len(selected) < limit {
		bestIdx := 0
		bestScore := math.Inf(-1)
		for idx, candidate := range remaining {
			redundancy := 0.0
			for _, chosen := range selected {
				var similarity float64
				if len(candidate.Embedding) > 0 && len(chosen.Embedding) > 0 {
					similarity = cosineSimilarity(candidate.Embedding, chosen.Embedding)
				} else {
					similarity = jaccardSimilarity(candidate.Normalized, chosen.Normalized)
				}
				if similarity > redundancy {
					redundancy = similarity
				}
			}
			score := 0.7*candidate.CombinedScore - 0.3*redundancy
			if candidate.Item.Pinned {
				score += 0.25
			}
			if score > bestScore {
				bestScore = score
				bestIdx = idx
			}
		}
		selected = append(selected, remaining[bestIdx])
		remaining = append(remaining[:bestIdx], remaining[bestIdx+1:]...)
	}
	return selected
}

func utf8RuneCount(s string) int {
	return len([]rune(s))
}

func buildFTSMatchQuery(query string) string {
	fragments := splitFTSFragments(query)
	if len(fragments) == 0 {
		return ""
	}

	clauses := make([]string, 0, len(fragments)*2)
	seen := map[string]bool{}
	addClause := func(value string) {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			return
		}
		seen[value] = true
		clauses = append(clauses, `"`+value+`"`)
	}

	for _, fragment := range fragments {
		runes := []rune(fragment)
		if len(runes) <= 8 {
			addClause(fragment)
		} else {
			addClause(string(runes[:8]))
			addClause(string(runes[len(runes)-8:]))
		}
		for _, trigram := range buildRuneNGrams(fragment, 3, 12) {
			addClause(trigram)
		}
		if len(clauses) >= 24 {
			break
		}
	}

	if len(clauses) == 0 {
		return ""
	}
	return strings.Join(clauses, " OR ")
}

func splitFTSFragments(query string) []string {
	query = strings.ReplaceAll(query, `"`, " ")
	parts := strings.FieldsFunc(query, func(r rune) bool {
		return unicode.IsSpace(r) || unicode.IsPunct(r) || unicode.IsSymbol(r)
	})
	out := make([]string, 0, len(parts))
	seen := map[string]bool{}
	for _, part := range parts {
		part = NormalizeText(part)
		if utf8RuneCount(part) < 3 || seen[part] {
			continue
		}
		seen[part] = true
		out = append(out, part)
		if len(out) >= 8 {
			break
		}
	}
	return out
}

func buildRuneNGrams(text string, size int, limit int) []string {
	runes := []rune(text)
	if len(runes) < size {
		if len(runes) == 0 {
			return nil
		}
		return []string{text}
	}
	out := make([]string, 0, len(runes)-size+1)
	seen := map[string]bool{}
	for i := 0; i <= len(runes)-size; i++ {
		part := string(runes[i : i+size])
		if seen[part] {
			continue
		}
		seen[part] = true
		out = append(out, part)
		if limit > 0 && len(out) >= limit {
			break
		}
	}
	return out
}

func EnqueueJob(jobType string, uniqueKey string, payload any) error {
	if db.SQLDB() == nil {
		return errors.New("database is not started")
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	now := nowString()
	err = sqlcgen.New(db.SQLDB()).InsertMemoryJobIgnoreConflict(context.Background(), sqlcgen.InsertMemoryJobIgnoreConflictParams{
		Type:        jobType,
		Status:      jobStatusQueued,
		UniqueKey:   uniqueKey,
		PayloadJson: string(body),
		Attempts:    0,
		AvailableAt: now,
		CreatedAt:   now,
		UpdatedAt:   now,
		LastError:   "",
	})
	return err
}

func EnqueueExtractTurn(payload ExtractTurnPayload) error {
	key := fmt.Sprintf("extract:%s:%s:%s:%d", payload.OwnerID, payload.CharacterID, payload.SessionID, payload.AssistantIndex)
	return EnqueueJob(jobTypeExtractTurn, key, payload)
}

func EnqueueCompactSession(payload CompactSessionPayload) error {
	key := fmt.Sprintf("compact:%s:%s:%s", payload.OwnerID, payload.CharacterID, payload.SessionID)
	return EnqueueJob(jobTypeCompactSession, key, payload)
}

func EnqueueReindexMemoryItem(memoryID string) error {
	return EnqueueJob(jobTypeReindexMemoryItem, "reindex:"+memoryID, ReindexMemoryItemPayload{MemoryID: memoryID})
}

func EnqueueRebuildCharacterMemory(ownerID string, characterID string) error {
	return EnqueueJob(jobTypeRebuildCharacter, fmt.Sprintf("rebuild:%s:%s", ownerID, characterID), RebuildCharacterMemoryPayload{
		OwnerID:     ownerID,
		CharacterID: characterID,
	})
}

func recoverRunningJobs() error {
	if db.SQLDB() == nil {
		return errors.New("database is not started")
	}
	return sqlcgen.New(db.SQLDB()).RecoverRunningMemoryJobs(context.Background(), sqlcgen.RecoverRunningMemoryJobsParams{
		Status:         jobStatusQueued,
		UpdatedAt:      nowString(),
		PreviousStatus: jobStatusRunning,
	})
}

func claimNextJob(ctx context.Context) (Job, bool, error) {
	conn := db.SQLDB()
	if conn == nil {
		return Job{}, false, errors.New("database is not started")
	}
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return Job{}, false, err
	}
	defer tx.Rollback()

	dbq := sqlcgen.New(tx)
	row, err := dbq.GetNextQueuedMemoryJob(ctx, sqlcgen.GetNextQueuedMemoryJobParams{
		Status:      jobStatusQueued,
		AvailableAt: nowString(),
	})
	var job Job
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Job{}, false, nil
		}
		return Job{}, false, err
	}
	job = Job{
		ID:         row.ID,
		Type:       row.Type,
		UniqueKey:  row.UniqueKey,
		PayloadRaw: row.PayloadJson,
		Attempts:   row.Attempts,
	}
	if err := dbq.MarkMemoryJobRunning(ctx, sqlcgen.MarkMemoryJobRunningParams{
		Status:    jobStatusRunning,
		UpdatedAt: nowString(),
		ID:        job.ID,
	}); err != nil {
		return Job{}, false, err
	}
	if err := tx.Commit(); err != nil {
		return Job{}, false, err
	}
	job.Attempts++
	return job, true, nil
}

func completeJob(ctx context.Context, id int64) error {
	if db.SQLDB() == nil {
		return errors.New("database is not started")
	}
	return sqlcgen.New(db.SQLDB()).DeleteMemoryJob(ctx, id)
}

func failJob(ctx context.Context, id int64, attempts int64, err error) error {
	if db.SQLDB() == nil {
		return errors.New("database is not started")
	}
	status := jobStatusQueued
	availableAt := timeNow().Add(defaultWorkerRetryBaseDelay * time.Duration(attempts*attempts)).UTC().Format(time.RFC3339)
	if attempts >= 3 {
		status = jobStatusFailed
	}
	return sqlcgen.New(db.SQLDB()).UpdateMemoryJobFailure(ctx, sqlcgen.UpdateMemoryJobFailureParams{
		Status:      status,
		AvailableAt: availableAt,
		UpdatedAt:   nowString(),
		LastError:   err.Error(),
		ID:          id,
	})
}

func listMemoryRecords(ownerID string, characterID string) ([]memoryItemRecord, error) {
	if db.SQLDB() == nil {
		return nil, errors.New("database is not started")
	}
	rows, err := sqlcgen.New(db.SQLDB()).ListMemoryRecordsForCharacter(context.Background(), sqlcgen.ListMemoryRecordsForCharacterParams{
		CharacterID: characterID,
		OwnerID:     ownerID,
	})
	if err != nil {
		return nil, err
	}
	out := make([]memoryItemRecord, 0, len(rows))
	for _, row := range rows {
		out = append(out, recordFromSQLC(row))
	}
	return out, nil
}

func boolToInt(v bool) int64 {
	if v {
		return 1
	}
	return 0
}
