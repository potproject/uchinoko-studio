package db

import "github.com/potproject/uchinoko-studio/data"

type chatSessionRow struct {
	SessionID    string `db:"session_id"`
	CharacterID  string `db:"character_id"`
	MessagesJSON string `db:"messages_json"`
}

func (r chatSessionRow) toMessage() (data.ChatMessage, error) {
	messages, err := unmarshalJSONString[[]data.ChatCompletionMessage](r.MessagesJSON)
	if err != nil {
		return data.ChatMessage{}, err
	}
	return data.ChatMessage{Chat: messages}, nil
}

func newChatSessionRow(id string, characterID string, message data.ChatMessage) (chatSessionRow, error) {
	messagesJSON, err := marshalJSONString(message.Chat)
	if err != nil {
		return chatSessionRow{}, err
	}

	return chatSessionRow{
		SessionID:    id,
		CharacterID:  characterID,
		MessagesJSON: messagesJSON,
	}, nil
}

func initChatMessage() data.ChatMessage {
	return data.ChatMessage{
		Chat: []data.ChatCompletionMessage{},
	}
}

func PutChatMessage(id string, characterId string, cM data.ChatMessage) error {
	row, err := newChatSessionRow(id, characterId, cM)
	if err != nil {
		return err
	}

	_, err = db.NamedExec(`
		INSERT INTO chat_sessions (session_id, character_id, messages_json)
		VALUES (:session_id, :character_id, :messages_json)
		ON CONFLICT(session_id, character_id) DO UPDATE SET
			messages_json = excluded.messages_json
	`, row)
	return err
}

func GetChatMessage(id string, characterId string) (cm data.ChatMessage, empty bool, err error) {
	var row chatSessionRow
	err = db.Get(&row, "SELECT * FROM chat_sessions WHERE session_id = ? AND character_id = ?", id, characterId)
	if isNotFound(err) {
		return initChatMessage(), true, nil
	}
	if err != nil {
		return data.ChatMessage{}, false, err
	}

	message, err := row.toMessage()
	if err != nil {
		return data.ChatMessage{}, false, err
	}
	return message, false, nil
}

func DeleteChatMessage(id string, characterId string) error {
	_, err := db.Exec("DELETE FROM chat_sessions WHERE session_id = ? AND character_id = ?", id, characterId)
	return err
}
