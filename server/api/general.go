package api

type TextMessage struct {
	Text    string
	IsFinal bool
}

type AudioMessage struct {
	Audio   []byte
	Text    string
	IsFinal bool
}
