package api

import (
	"context"
	"errors"
	"fmt"

	speech "cloud.google.com/go/speech/apiv1"
	"cloud.google.com/go/speech/apiv1/speechpb"
	"google.golang.org/api/option"
)

func GoSpeech(apiKey string, fileData []byte, extension string, language string) (string, error) {
	ctx := context.Background()

	client, err := speech.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return "", fmt.Errorf("failed to create speech client: %v", err)
	}
	defer client.Close()

	// Determine the correct audio encoding format
	var encoding speechpb.RecognitionConfig_AudioEncoding
	switch extension {
	case "wav":
		encoding = speechpb.RecognitionConfig_LINEAR16
	case "mp3":
		encoding = speechpb.RecognitionConfig_MP3
	case "ogg":
		encoding = speechpb.RecognitionConfig_OGG_OPUS
	case "webm":
		encoding = speechpb.RecognitionConfig_WEBM_OPUS
	default:
		return "", errors.New("unsupported audio format")
	}

	config := &speechpb.RecognitionConfig{
		Encoding:     encoding,
		LanguageCode: language,
	}

	// Set up the request
	req := &speechpb.RecognizeRequest{
		Config: config,
		Audio: &speechpb.RecognitionAudio{
			AudioSource: &speechpb.RecognitionAudio_Content{
				Content: fileData,
			},
		},
	}

	// Perform the request
	resp, err := client.Recognize(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to recognize: %v", err)
	}

	// Process the response and return the transcribed text
	var transcript string
	for _, result := range resp.Results {
		for _, alt := range result.Alternatives {
			transcript += alt.Transcript
		}
	}

	return transcript, nil
}
