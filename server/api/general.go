package api

import "github.com/potproject/uchinoko-studio/data"

type TextChunkMessage struct {
	Text  string
	Voice data.CharacterConfigVoice
}

// type BehaviorChunkMessage struct {
// 	Behavior string
// }

type ChunkMessage interface {
	// TextMessage or BehaviorChunkMessage
}

type AudioMessage struct {
	Audio *[]byte
	Text  string
	Voice data.CharacterConfigVoice
}
