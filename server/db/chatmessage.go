package db

import (
	"encoding/json"

	"github.com/potproject/uchinoko/data"
	"github.com/sashabaranov/go-openai"
	"github.com/syndtr/goleveldb/leveldb"
)

func initChatMessage() data.ChatMessage {
	return data.ChatMessage{
		Chat: []openai.ChatCompletionMessage{},
	}
}

func PutChatMessage(id string, cM data.ChatMessage) error {
	key := []byte(id + "/chatmessage")
	value, err := json.Marshal(cM)
	if err != nil {
		return err
	}
	return put(key, value)
}

func GetChatMessage(id string) (data.ChatMessage, error) {
	key := []byte(id + "/chatmessage")
	value, err := get(key)
	if err == leveldb.ErrNotFound {
		return initChatMessage(), nil
	} else if err != nil {
		return data.ChatMessage{}, err
	}
	var cM data.ChatMessage
	err = json.Unmarshal(value, &cM)
	if err != nil {
		return data.ChatMessage{}, err
	}
	return cM, nil
}
