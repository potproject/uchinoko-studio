package api

import "github.com/potproject/uchinoko-studio/data"

type TextMessage struct {
	Text  string
	Voice data.CharacterConfigVoice
}

type AudioMessage struct {
	Audio *[]byte
	Text  string
	Voice data.CharacterConfigVoice
}
