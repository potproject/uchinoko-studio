package texttospeech

import (
	"context"
	"fmt"

	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	texttospeechpb "cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
	"google.golang.org/api/option"
)

func goSpeech(apiKey string, language string, voiceType string, voiceName string, text string) ([]byte, error) {
	ctx := context.Background()

	client, err := texttospeech.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %v", err)
	}
	defer client.Close()

	req := &texttospeechpb.SynthesizeSpeechRequest{
		Input: &texttospeechpb.SynthesisInput{
			InputSource: &texttospeechpb.SynthesisInput_Text{Text: text},
		},
		Voice: &texttospeechpb.VoiceSelectionParams{
			LanguageCode: language,
			SsmlGender:   getGender(voiceType),
			Name:         voiceName,
		},
		AudioConfig: &texttospeechpb.AudioConfig{
			AudioEncoding: texttospeechpb.AudioEncoding_LINEAR16,
		},
	}

	resp, err := client.SynthesizeSpeech(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to synthesize speech: %v", err)
	}

	return resp.AudioContent, nil
}

func getGender(voiceType string) texttospeechpb.SsmlVoiceGender {
	switch voiceType {
	case "MALE":
		return texttospeechpb.SsmlVoiceGender_MALE
	case "FEMALE":
		return texttospeechpb.SsmlVoiceGender_FEMALE
	case "NEUTRAL":
		return texttospeechpb.SsmlVoiceGender_NEUTRAL
	default:
		return texttospeechpb.SsmlVoiceGender_SSML_VOICE_GENDER_UNSPECIFIED
	}
}
