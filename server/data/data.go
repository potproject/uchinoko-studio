package data

import "github.com/sashabaranov/go-openai"

type Config struct {
	Name string `json:"name"`
}

type ChatMessage struct {
	Chat []openai.ChatCompletionMessage
}
