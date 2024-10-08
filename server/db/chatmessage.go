package db

import (
	_ "embed"
	"encoding/json"

	"github.com/potproject/uchinoko-studio/data"
	"github.com/syndtr/goleveldb/leveldb"
)

func initChatMessage() data.ChatMessage {
	return data.ChatMessage{
		Chat: []data.ChatCompletionMessage{},
	}
}

func PutChatMessage(id string, characterId string, cM data.ChatMessage) error {
	key := []byte(id + "/" + characterId + "/chatmessage")
	value, err := json.Marshal(cM)
	if err != nil {
		return err
	}
	return put(key, value)
}

func GetChatMessage(id string, characterId string) (cm data.ChatMessage, empty bool, err error) {
	key := []byte(id + "/" + characterId + "/chatmessage")
	value, err := get(key)
	if err == leveldb.ErrNotFound {
		return initChatMessage(), true, nil
	} else if err != nil {
		return data.ChatMessage{}, false, err
	}
	var cM data.ChatMessage
	err = json.Unmarshal(value, &cM)
	if err != nil {
		return data.ChatMessage{}, false, err
	}
	return cM, false, nil
}

func DeleteChatMessage(id string, characterId string) error {
	key := []byte(id + "/" + characterId + "/chatmessage")
	return delete(key)
}
