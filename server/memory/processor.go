package memory

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/potproject/uchinoko-studio/data"
	"github.com/potproject/uchinoko-studio/db"
	"github.com/potproject/uchinoko-studio/db/sqlcgen"
	"github.com/potproject/uchinoko-studio/prompts"
)

type extractionResponse struct {
	Noop    bool                  `json:"noop"`
	Upserts []extractedMemoryItem `json:"upserts"`
}

type extractedMemoryItem struct {
	Kind         string  `json:"kind"`
	Content      string  `json:"content"`
	KeywordsText string  `json:"keywordsText"`
	Confidence   float64 `json:"confidence"`
	Salience     float64 `json:"salience"`
	Pinned       bool    `json:"pinned"`
}

func processJob(ctx context.Context, job Job) error {
	switch job.Type {
	case jobTypeExtractTurn:
		var payload ExtractTurnPayload
		if err := json.Unmarshal([]byte(job.PayloadRaw), &payload); err != nil {
			return err
		}
		return processExtractTurn(ctx, payload)
	case jobTypeCompactSession:
		var payload CompactSessionPayload
		if err := json.Unmarshal([]byte(job.PayloadRaw), &payload); err != nil {
			return err
		}
		return processCompactSession(ctx, payload)
	case jobTypeReindexMemoryItem:
		var payload ReindexMemoryItemPayload
		if err := json.Unmarshal([]byte(job.PayloadRaw), &payload); err != nil {
			return err
		}
		return processReindexMemoryItem(ctx, payload)
	case jobTypeRebuildCharacter:
		var payload RebuildCharacterMemoryPayload
		if err := json.Unmarshal([]byte(job.PayloadRaw), &payload); err != nil {
			return err
		}
		return processRebuildCharacterMemory(ctx, payload)
	default:
		return fmt.Errorf("unsupported memory job type: %s", job.Type)
	}
}

func processExtractTurn(ctx context.Context, payload ExtractTurnPayload) error {
	character, err := db.GetCharacterConfig(payload.CharacterID)
	if err != nil {
		return err
	}
	if !character.Memory.Enabled || !character.Memory.EnableRelationshipMemory {
		return nil
	}

	systemPrompt := prompts.MemoryExtractTurnSystem
	userPrompt := fmt.Sprintf("ユーザー発話:\n%s\n\nアシスタント応答:\n%s", payload.UserContent, payload.AssistantContent)
	raw, err := completeText(character, systemPrompt, userPrompt)
	if err != nil {
		return err
	}
	resp := extractionResponse{Noop: true}
	if err := decodeLooseJSON(raw, &resp); err != nil {
		return nil
	}
	if resp.Noop || len(resp.Upserts) == 0 {
		return nil
	}
	reindexIDs := make([]string, 0, len(resp.Upserts))
	err = withTx(ctx, func(tx *sql.Tx) error {
		for _, item := range resp.Upserts {
			if strings.TrimSpace(item.Content) == "" {
				continue
			}
			if !isAllowedRelationshipKind(item.Kind) {
				continue
			}
			created, rowID, err := upsertRelationshipMemoryTx(tx, payload.OwnerID, payload.CharacterID, item)
			if err != nil {
				return err
			}
			if err := insertEvidenceTx(tx, rowID, payload.SessionID, payload.UserIndex, payload.AssistantIndex, payload.UserContent+"\n"+payload.AssistantContent); err != nil {
				return err
			}
			reindexIDs = append(reindexIDs, created.ID)
		}
		return nil
	})
	if err != nil {
		return err
	}
	for _, id := range reindexIDs {
		_ = EnqueueReindexMemoryItem(id)
	}
	return nil
}

func processCompactSession(ctx context.Context, payload CompactSessionPayload) error {
	if payload.MaxHistory <= 0 {
		return nil
	}
	character, err := db.GetCharacterConfig(payload.CharacterID)
	if err != nil {
		return err
	}
	if !character.Memory.Enabled {
		return nil
	}

	message, _, err := db.GetChatMessage(payload.SessionID, payload.CharacterID)
	if err != nil {
		return err
	}
	if int64(len(message.Chat)) <= payload.MaxHistory {
		return nil
	}

	keepStart := len(message.Chat) - int(payload.MaxHistory)
	if keepStart <= 0 {
		return nil
	}
	compactMessages := append([]data.ChatCompletionMessage(nil), message.Chat[:keepStart]...)
	remaining := append([]data.ChatCompletionMessage(nil), message.Chat[keepStart:]...)
	summary, err := GetSessionSummary(payload.OwnerID, payload.CharacterID, payload.SessionID)
	if err != nil {
		return err
	}

	summaryPrompt := prompts.MemoryCompactSummarySystem
	userPrompt := fmt.Sprintf("既存 summary:\n%s\n\n今回圧縮する会話:\n%s", summary.Summary, renderMessages(compactMessages))
	newSummary, err := completeText(character, summaryPrompt, userPrompt)
	if err != nil {
		return err
	}

	extractPrompt := prompts.MemoryCompactExtractSystem
	rawExtraction, err := completeText(character, extractPrompt, renderMessages(compactMessages))
	if err != nil {
		return err
	}
	extracted := extractionResponse{Noop: true}
	_ = decodeLooseJSON(rawExtraction, &extracted)

	reindexIDs := make([]string, 0, len(extracted.Upserts))
	err = withTx(ctx, func(tx *sql.Tx) error {
		if character.Memory.EnableSessionSummary {
			if err := upsertSessionSummaryTx(tx, data.SessionSummary{
				OwnerID:     payload.OwnerID,
				CharacterID: payload.CharacterID,
				SessionID:   payload.SessionID,
				Summary:     strings.TrimSpace(newSummary),
				UpdatedAt:   nowString(),
			}); err != nil {
				return err
			}
		}

		for _, item := range extracted.Upserts {
			if !isAllowedRelationshipKind(item.Kind) || strings.TrimSpace(item.Content) == "" {
				continue
			}
			created, rowID, err := upsertRelationshipMemoryTx(tx, payload.OwnerID, payload.CharacterID, item)
			if err != nil {
				return err
			}
			if err := insertEvidenceTx(tx, rowID, payload.SessionID, 0, int64(keepStart-1), renderMessages(compactMessages)); err != nil {
				return err
			}
			reindexIDs = append(reindexIDs, created.ID)
		}
		return replaceChatMessagesTx(tx, payload.SessionID, payload.CharacterID, remaining)
	})
	if err != nil {
		return err
	}
	for _, id := range reindexIDs {
		_ = EnqueueReindexMemoryItem(id)
	}
	return nil
}

func processReindexMemoryItem(ctx context.Context, payload ReindexMemoryItemPayload) error {
	item, rowID, err := getMemoryItemByPublicID(payload.MemoryID)
	if err != nil {
		return err
	}
	character, err := db.GetCharacterConfig(item.CharacterID)
	if err != nil {
		return err
	}
	embedding := []float64(nil)
	if character.Memory.Enabled && character.Memory.EnableSemanticSearch {
		embedding, _ = embedText(item.Content+"\n"+item.KeywordsText, character.Memory)
	}
	encodedEmbedding := ""
	if len(embedding) > 0 {
		body, _ := json.Marshal(embedding)
		encodedEmbedding = string(body)
	}
	return sqlcgen.New(db.SQLDB()).UpdateMemoryItemIndex(ctx, sqlcgen.UpdateMemoryItemIndexParams{
		NormalizedContent: NormalizeText(item.Content),
		KeywordsText:      NormalizeText(item.KeywordsText),
		EmbeddingJson:     encodedEmbedding,
		UpdatedAt:         nowString(),
		ID:                rowID,
	})
}

func processRebuildCharacterMemory(ctx context.Context, payload RebuildCharacterMemoryPayload) error {
	records, err := listMemoryRecords(payload.OwnerID, payload.CharacterID)
	if err != nil {
		return err
	}
	sort.Slice(records, func(i, j int) bool { return records[i].rowID < records[j].rowID })
	for _, record := range records {
		if err := processReindexMemoryItem(ctx, ReindexMemoryItemPayload{MemoryID: record.ID}); err != nil {
			return err
		}
	}
	return nil
}

func upsertRelationshipMemoryTx(tx *sql.Tx, ownerID string, characterID string, extracted extractedMemoryItem) (data.MemoryItem, int64, error) {
	now := nowString()
	normalized := NormalizeText(extracted.Content)
	keywords := NormalizeText(extracted.KeywordsText)

	rows, err := sqlcgen.New(tx).ListActiveRelationshipMemory(context.Background(), sqlcgen.ListActiveRelationshipMemoryParams{
		CharacterID: characterID,
		OwnerID:     ownerID,
	})
	if err != nil {
		return data.MemoryItem{}, 0, err
	}

	type existing struct {
		rowID    int64
		publicID string
		content  string
		kind     string
	}
	candidates := make([]existing, 0, len(rows))
	for _, row := range rows {
		candidates = append(candidates, existing{
			rowID:    row.ID,
			publicID: row.PublicID,
			content:  row.Content,
			kind:     row.Kind,
		})
	}

	publicID := uuid.New().String()
	res, err := sqlcgen.New(tx).InsertMemoryItem(context.Background(), sqlcgen.InsertMemoryItemParams{
		PublicID:          publicID,
		CharacterID:       characterID,
		OwnerID:           ownerID,
		Scope:             string(data.MemoryScopeRelationship),
		Kind:              extracted.Kind,
		Content:           extracted.Content,
		NormalizedContent: normalized,
		KeywordsText:      keywords,
		Pinned:            boolToInt(extracted.Pinned),
		Confidence:        extracted.Confidence,
		Salience:          extracted.Salience,
		Source:            "extracted",
		SupersededBy:      sql.NullInt64{},
		EmbeddingJson:     "",
		CreatedAt:         now,
		UpdatedAt:         now,
		LastAccessedAt:    now,
	})
	if err != nil {
		return data.MemoryItem{}, 0, err
	}
	rowID, err := res.LastInsertId()
	if err != nil {
		return data.MemoryItem{}, 0, err
	}

	for _, candidate := range candidates {
		if candidate.kind != extracted.Kind {
			continue
		}
		if jaccardSimilarity(NormalizeText(candidate.content), normalized) >= 0.72 {
			if err := sqlcgen.New(tx).SupersedeMemoryItem(context.Background(), sqlcgen.SupersedeMemoryItemParams{
				SupersededBy: sql.NullInt64{Int64: rowID, Valid: true},
				ID:           candidate.rowID,
			}); err != nil {
				return data.MemoryItem{}, 0, err
			}
		}
	}
	return data.MemoryItem{
		ID:           publicID,
		CharacterID:  characterID,
		OwnerID:      ownerID,
		Scope:        data.MemoryScopeRelationship,
		Kind:         extracted.Kind,
		Content:      extracted.Content,
		KeywordsText: extracted.KeywordsText,
		Pinned:       extracted.Pinned,
		Confidence:   extracted.Confidence,
		Salience:     extracted.Salience,
		Source:       "extracted",
		UpdatedAt:    now,
	}, rowID, nil
}

func insertEvidenceTx(tx *sql.Tx, memoryRowID int64, sessionID string, start int64, end int64, excerpt string) error {
	return sqlcgen.New(tx).InsertMemoryItemEvidence(context.Background(), sqlcgen.InsertMemoryItemEvidenceParams{
		MemoryItemID:      memoryRowID,
		SessionID:         sessionID,
		MessageIndexStart: start,
		MessageIndexEnd:   end,
		Excerpt:           truncateString(excerpt, 2000),
		CreatedAt:         nowString(),
	})
}

func isAllowedRelationshipKind(kind string) bool {
	switch kind {
	case "preference", "profile_fact", "relationship_fact", "promise", "ongoing_topic", "instruction_preference":
		return true
	default:
		return false
	}
}

func renderMessages(messages []data.ChatCompletionMessage) string {
	var lines []string
	for _, message := range messages {
		role := "user"
		if message.Role == data.ChatCompletionMessageRoleAssistant {
			role = "assistant"
		}
		lines = append(lines, fmt.Sprintf("%s: %s", role, strings.TrimSpace(message.Content)))
	}
	return strings.Join(lines, "\n")
}

func truncateString(value string, max int) string {
	runes := []rune(value)
	if len(runes) <= max {
		return value
	}
	return string(runes[:max])
}

func decodeLooseJSON(raw string, out any) error {
	raw = strings.TrimSpace(raw)
	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start < 0 || end <= start {
		return fmt.Errorf("json object not found")
	}
	return json.Unmarshal([]byte(raw[start:end+1]), out)
}

func replaceChatMessagesTx(tx *sql.Tx, sessionID string, characterID string, messages []data.ChatCompletionMessage) error {
	qtx := sqlcgen.New(tx)
	if err := qtx.UpsertChatSession(context.Background(), sqlcgen.UpsertChatSessionParams{
		SessionID:   sessionID,
		CharacterID: characterID,
	}); err != nil {
		return err
	}
	if err := qtx.DeleteChatMessages(context.Background(), sqlcgen.DeleteChatMessagesParams{
		SessionID:   sessionID,
		CharacterID: characterID,
	}); err != nil {
		return err
	}
	for idx, message := range messages {
		imageExtension := ""
		var imageData []byte
		if message.Image != nil {
			imageExtension = message.Image.Extension
			imageData = message.Image.Data
		}
		if err := qtx.InsertChatMessage(context.Background(), sqlcgen.InsertChatMessageParams{
			SessionID:      sessionID,
			CharacterID:    characterID,
			MessageIndex:   int64(idx),
			Role:           message.Role,
			Content:        message.Content,
			ImageExtension: imageExtension,
			ImageData:      imageData,
		}); err != nil {
			return err
		}
	}
	return nil
}
