package chat

import (
	"strings"
	"unicode/utf8"

	"github.com/potproject/uchinoko-studio/api"
	"github.com/potproject/uchinoko-studio/data"
)

const chars = ".,?!;:—-)]} 。、？！；：」）］｝　\"'"

type ChatStream func(string, []data.CharacterConfigVoice, bool, string, string, []data.ChatCompletionMessage, string, chan api.ChunkMessage) ([]data.ChatCompletionMessage, error)

func chatReceiver(
	charChannel chan rune,
	done chan bool,
	multi bool,
	voices []data.CharacterConfigVoice,
	chunkMessage chan api.ChunkMessage,
	text string,
	cm []data.ChatCompletionMessage,
) ([]data.ChatCompletionMessage, error) {
	voice := voices[0]
	voiceIndentifications := make([]string, len(voices))
	if multi {
		for i, v := range voices {
			voiceIndentifications[i] = v.Identification
		}
	}

	allText := ""
	bufferText := ""
	for {
		select {
		case c := <-charChannel:
			allText += string(c)
			bufferText += string(c)
			if len(voice.Behavior) > 0 {
				for _, v := range voice.Behavior {
					if strings.Contains(bufferText, v.Identification) {
						bufferText = strings.Replace(bufferText, v.Identification, "", -1)
						chunkMessage <- api.BehaviorChunkMessage{
							Behavior: v,
						}
						break
					}
				}
			}

			if multi {
				for i, v := range voiceIndentifications {
					if strings.Contains(bufferText, v) {
						bufferText = strings.Replace(bufferText, v, "", -1)
						voice = voices[i]
						break
					}
				}
			}

			contain := strings.Contains(chars, string(c))
			if contain && utf8.RuneCountInString(bufferText) > 1 {
				chunkMessage <- api.TextChunkMessage{
					Text:  bufferText,
					Voice: voice,
				}
				bufferText = ""
			}
		case <-done:
			chunkMessage <- api.TextChunkMessage{
				Text:  bufferText,
				Voice: voice,
			}
			return append(
				cm,
				data.ChatCompletionMessage{
					Role:    data.ChatCompletionMessageRoleUser,
					Content: text,
				},
				data.ChatCompletionMessage{
					Role:    data.ChatCompletionMessageRoleAssistant,
					Content: strings.Trim(allText, "\n"),
				},
			), nil
		}
	}
}