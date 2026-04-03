package db

import (
	"context"
	"strings"

	"github.com/potproject/uchinoko-studio/data"
	"github.com/potproject/uchinoko-studio/db/sqlcgen"
)

func loadChatMessage(ctx context.Context, q *sqlcgen.Queries, row sqlcgen.ChatSession) (data.ChatMessage, error) {
	messageRows, err := q.ListChatMessages(ctx, sqlcgen.ListChatMessagesParams{
		SessionID:   row.SessionID,
		CharacterID: row.CharacterID,
	})
	if err != nil {
		return data.ChatMessage{}, err
	}

	messages := make([]data.ChatCompletionMessage, 0, len(messageRows))
	for _, messageRow := range messageRows {
		var image *data.Image
		if messageRow.ImageExtension != "" {
			image = &data.Image{
				Extension: messageRow.ImageExtension,
				Data:      append([]byte(nil), messageRow.ImageData...),
			}
		}
		messages = append(messages, data.ChatCompletionMessage{
			Role:    messageRow.Role,
			Content: messageRow.Content,
			Image:   image,
		})
	}

	return data.ChatMessage{Chat: messages}, nil
}

func newChatSessionParams(id string, characterID string) sqlcgen.UpsertChatSessionParams {
	return sqlcgen.UpsertChatSessionParams{
		SessionID:   id,
		CharacterID: characterID,
	}
}

func initChatMessage() data.ChatMessage {
	return data.ChatMessage{
		Chat: []data.ChatCompletionMessage{},
	}
}

func PutChatMessage(id string, characterId string, cM data.ChatMessage) error {
	ctx := context.Background()

	return withTx(ctx, func(qtx *sqlcgen.Queries) error {
		if err := qtx.UpsertChatSession(ctx, newChatSessionParams(id, characterId)); err != nil {
			return err
		}
		if err := qtx.DeleteChatMessages(ctx, sqlcgen.DeleteChatMessagesParams{
			SessionID:   id,
			CharacterID: characterId,
		}); err != nil {
			return err
		}

		for messageIndex, message := range cM.Chat {
			imageExtension := ""
			var imageData []byte
			if message.Image != nil {
				imageExtension = message.Image.Extension
				imageData = message.Image.Data
			}

			if err := qtx.InsertChatMessage(ctx, sqlcgen.InsertChatMessageParams{
				SessionID:      id,
				CharacterID:    characterId,
				MessageIndex:   int64(messageIndex),
				Role:           message.Role,
				Content:        message.Content,
				ImageExtension: imageExtension,
				ImageData:      imageData,
			}); err != nil {
				return err
			}
		}

		return nil
	})
}

func GetChatMessage(id string, characterId string) (cm data.ChatMessage, empty bool, err error) {
	row, err := queries.GetChatSession(context.Background(), sqlcgen.GetChatSessionParams{
		SessionID:   id,
		CharacterID: characterId,
	})
	if isNotFound(err) {
		return initChatMessage(), true, nil
	}
	if err != nil {
		return data.ChatMessage{}, false, err
	}

	message, err := loadChatMessage(context.Background(), queries, row)
	if err != nil {
		return data.ChatMessage{}, false, err
	}
	return message, false, nil
}

func DeleteChatMessage(id string, characterId string) error {
	return queries.DeleteChatSession(context.Background(), sqlcgen.DeleteChatSessionParams{
		SessionID:   id,
		CharacterID: characterId,
	})
}

func ListChatSessionsForOwner(ownerID string, characterID string) (data.ChatSessionList, error) {
	rows, err := queries.ListChatSessionsByCharacter(context.Background(), characterID)
	if err != nil {
		return data.ChatSessionList{}, err
	}

	sessions := make([]data.ChatSessionSummary, 0, len(rows))
	for _, row := range rows {
		if !isOwnedChatSession(ownerID, row.SessionID) {
			continue
		}

		message, err := loadChatMessage(context.Background(), queries, row)
		if err != nil {
			return data.ChatSessionList{}, err
		}

		sessions = append(sessions, data.ChatSessionSummary{
			SessionID:    row.SessionID,
			Title:        deriveChatSessionTitle(message),
			Preview:      deriveChatSessionPreview(message),
			MessageCount: len(message.Chat),
			IsDefault:    row.SessionID == ownerID,
		})
	}

	return data.ChatSessionList{Sessions: sessions}, nil
}

func isOwnedChatSession(ownerID string, sessionID string) bool {
	if sessionID == ownerID {
		return true
	}
	return strings.HasPrefix(sessionID, ownerID+":")
}

func deriveChatSessionTitle(message data.ChatMessage) string {
	for _, item := range message.Chat {
		content := strings.TrimSpace(item.Content)
		if content == "" {
			continue
		}
		if item.Role == data.ChatCompletionMessageRoleUser {
			return content
		}
	}
	for _, item := range message.Chat {
		content := strings.TrimSpace(item.Content)
		if content != "" {
			return content
		}
	}
	return "新しいチャット"
}

func deriveChatSessionPreview(message data.ChatMessage) string {
	for index := len(message.Chat) - 1; index >= 0; index-- {
		content := strings.TrimSpace(message.Chat[index].Content)
		if content != "" {
			return content
		}
	}
	return ""
}
