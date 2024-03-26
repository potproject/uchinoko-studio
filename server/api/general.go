package api

type TextMessage struct {
	Text string
}

type AudioMessage struct {
	Audio []byte
	Text  string
}
