package api

import (
	"github.com/haguro/elevenlabs-go"
)

func ElevenLabsTTS(c *ElevenLabsClientExtend, voiceID string, text string) ([]byte, error) {
	bytes, err := c.Client.TextToSpeech(voiceID,
		elevenlabs.TextToSpeechRequest{
			Text:    text,
			ModelID: elevenlabsModelID,
		},
	)

	if err != nil {
		return nil, err
	}
	return bytes, nil
}
