package db

import (
	_ "embed"
	"encoding/json"

	"github.com/potproject/uchinoko-studio/data"
	"github.com/sashabaranov/go-openai"
	"github.com/syndtr/goleveldb/leveldb"
)

//go:embed propmt.txt
var systemMessage string

func initChatMessage() data.ChatMessage {
	return data.ChatMessage{
		Chat: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: systemMessage,
			},
		},
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

func GetChatMessage(id string) (cm data.ChatMessage, empty bool, err error) {
	key := []byte(id + "/chatmessage")
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
