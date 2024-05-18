package texttospeech

import (
	"log"
	"strings"

	"github.com/potproject/uchinoko-studio/api"
	"github.com/potproject/uchinoko-studio/data"
	"github.com/potproject/uchinoko-studio/envgen"
)

func getVoiceEndpoint(voiceType string) string {
	if voiceType == "voicevox" {
		return envgen.Get().VOICEVOX_ENDPOINT()
	}
	if voiceType == "bertvits2" {
		return envgen.Get().BERTVITS2_ENDPOINT()
	}
	if voiceType == "stylebertvits2" {
		return envgen.Get().STYLEBERTVIT2_ENDPOINT()
	}
	return ""
}

func getApiKey(voiceType string) string {
	if voiceType == "google-text-to-speech" {
		return envgen.Get().GOOGLE_TEXT_TO_SPEECH_API_KEY()
	}
	if voiceType == "openai-speech" {
		return envgen.Get().OPENAI_SPEECH_API_KEY()
	}
	return ""
}

func removeNewLineAndSpace(text string) string {
	return strings.ReplaceAll(strings.ReplaceAll(text, "\n", ""), " ", "")
}

func TTSStream(general data.GeneralConfig, chunkMessage <-chan api.ChunkMessage, changeVoice chan<- data.CharacterConfigVoice, changeBehavior chan<- data.CharacterConfigVoiceBehavior, outAudioMessage chan<- api.AudioMessage, chatDone chan bool) error {
	beforeVoiceIdentification := ""
	for {
		select {
		case c := <-chunkMessage:
			if t, ok := c.(api.TextChunkMessage); ok {
				escapeText := removeNewLineAndSpace(t.Text)
				if len(escapeText) == 0 {
					continue
				}

				if beforeVoiceIdentification != t.Voice.Identification {
					beforeVoiceIdentification = t.Voice.Identification
					changeVoice <- t.Voice
				}

				var bin []byte
				var err error

				if t.Voice.Type == "voicevox" {
					bin, err = voicevox(getVoiceEndpoint(t.Voice.Type), t.Voice.SpeakerID, t.Text)
				}
				if t.Voice.Type == "bertvits2" {
					bin, err = bertVits2(getVoiceEndpoint(t.Voice.Type), t.Voice.ModelID, t.Voice.SpeakerID, t.Text)
				}
				if t.Voice.Type == "stylebertvits2" {
					bin, err = styleBertVits2(getVoiceEndpoint(t.Voice.Type), t.Voice.ModelID, t.Voice.ModelFile, t.Voice.SpeakerID, t.Text)
				}
				if t.Voice.Type == "google-text-to-speech" {
					bin, err = goSpeech(getApiKey(t.Voice.Type), general.Language, t.Voice.SpeakerID, t.Voice.ModelID, t.Text)
				}
				if t.Voice.Type == "openai-speech" {
					bin, err = openAISpeech(getApiKey(t.Voice.Type), t.Voice.ModelID, t.Voice.SpeakerID, t.Text)
				}

				if err != nil {
					log.Printf("Error: %s", err.Error())
					return err
				}
				outAudioMessage <- api.AudioMessage{
					Audio: &bin,
					Text:  t.Text,
				}
			}
			if _, ok := c.(api.BehaviorChunkMessage); ok {
				changeBehavior <- c.(api.BehaviorChunkMessage).Behavior
			}
		case <-chatDone:
			return nil
		}
	}
}
