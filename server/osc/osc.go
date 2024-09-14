package osc

import (
	"errors"

	"github.com/hypebeast/go-osc/osc"
)

func SendChatBoxMessage(addr string, port int32, message string) error {
	client := osc.NewClient(addr, int(port))
	if client == nil {
		return errors.New("failed to create OSC client")
	}

	// 144文字以上のメッセージは送信できないため、丸める
	if len(message) > 144 {
		message = message[:140] + "..."
	}
	msg := osc.NewMessage("/chatbox/input")
	msg.Append(message)
	msg.Append(true)
	return client.Send(msg)
}
