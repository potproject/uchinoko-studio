package db

import (
	"context"

	"github.com/potproject/uchinoko-studio/data"
	"github.com/potproject/uchinoko-studio/db/sqlcgen"
)

func chatMessageFromRow(row sqlcgen.ChatSession) (data.ChatMessage, error) {
	messages, err := unmarshalJSONString[[]data.ChatCompletionMessage](row.MessagesJson)
	if err != nil {
		return data.ChatMessage{}, err
	}
	return data.ChatMessage{Chat: messages}, nil
}

func newChatSessionParams(id string, characterID string, message data.ChatMessage) (sqlcgen.UpsertChatSessionParams, error) {
	messagesJSON, err := marshalJSONString(message.Chat)
	if err != nil {
		return sqlcgen.UpsertChatSessionParams{}, err
	}

	return sqlcgen.UpsertChatSessionParams{
		SessionID:    id,
		CharacterID:  characterID,
		MessagesJson: messagesJSON,
	}, nil
}

func initChatMessage() data.ChatMessage {
	return data.ChatMessage{
		Chat: []data.ChatCompletionMessage{},
	}
}

func PutChatMessage(id string, characterId string, cM data.ChatMessage) error {
	row, err := newChatSessionParams(id, characterId, cM)
	if err != nil {
		return err
	}

	return queries.UpsertChatSession(context.Background(), row)
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

	message, err := chatMessageFromRow(row)
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
