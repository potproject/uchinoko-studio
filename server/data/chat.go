package data

const ChatCompletionMessageRoleAssistant = "assistant"
const ChatCompletionMessageRoleUser = "user"

type ChatMessage struct {
	Chat []ChatCompletionMessage
}

type ChatCompletionMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	Image   *Image
}

type Tokens struct {
	InputTokens  int64 `json:"input_tokens"`
	OutputTokens int64 `json:"output_tokens"`
}
