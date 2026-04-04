package controller

import (
	"testing"

	"github.com/gofiber/contrib/websocket"
)

func TestShouldIgnoreWhisperNoise(t *testing.T) {
	tests := []struct {
		name              string
		mt                int
		transcriptionType string
		requestText       string
		want              bool
	}{
		{
			name:              "ignores empty text",
			mt:                websocket.BinaryMessage,
			transcriptionType: "openai_speech_to_text",
			requestText:       "",
			want:              true,
		},
		{
			name:              "ignores japanese outro",
			mt:                websocket.BinaryMessage,
			transcriptionType: "openai_speech_to_text",
			requestText:       "ご視聴ありがとうございました",
			want:              true,
		},
		{
			name:              "ignores english thank you with punctuation",
			mt:                websocket.BinaryMessage,
			transcriptionType: "openai_speech_to_text",
			requestText:       " Thank you. ",
			want:              true,
		},
		{
			name:              "does not ignore longer sentence",
			mt:                websocket.BinaryMessage,
			transcriptionType: "openai_speech_to_text",
			requestText:       "Thank you for coming today",
			want:              false,
		},
		{
			name:              "does not ignore typed text",
			mt:                websocket.TextMessage,
			transcriptionType: "openai_speech_to_text",
			requestText:       "Thank you",
			want:              false,
		},
		{
			name:              "does not ignore other transcription providers",
			mt:                websocket.BinaryMessage,
			transcriptionType: "google_speech_to_text",
			requestText:       "Thank you",
			want:              false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldIgnoreWhisperNoise(tt.mt, tt.transcriptionType, tt.requestText)
			if got != tt.want {
				t.Fatalf("shouldIgnoreWhisperNoise() = %v, want %v", got, tt.want)
			}
		})
	}
}
