package api

import (
	"log"
	"strings"

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

func removeNewLineAndSpace(text string) string {
	return strings.ReplaceAll(strings.ReplaceAll(text, "\n", ""), " ", "")
}

func TTSStream(chunkMessage <-chan TextMessage, changeVoice chan<- data.CharacterConfigVoice, outAudioMessage chan<- AudioMessage, chatDone chan bool) error {
	beforeVoiceIdentification := ""
	for {
		select {
		case t := <-chunkMessage:
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
				bin, err = voicevoxTTS(getVoiceEndpoint(t.Voice.Type), t.Voice.SpeakerID, t.Text)
			}
			if t.Voice.Type == "bertvits2" {
				bin, err = bertVits2TTS(getVoiceEndpoint(t.Voice.Type), t.Voice.ModelID, t.Voice.SpeakerID, t.Text)
			}
			if t.Voice.Type == "stylebertvits2" {
				bin, err = styleBertVits2TTS(getVoiceEndpoint(t.Voice.Type), t.Voice.ModelID, t.Voice.ModelFile, t.Voice.SpeakerID, t.Text)
			}

			if err != nil {
				log.Printf("Error: %s", err.Error())
				return err
			}
			outAudioMessage <- AudioMessage{
				Audio: &bin,
				Text:  t.Text,
			}
		case <-chatDone:
			return nil
		}
	}
}
