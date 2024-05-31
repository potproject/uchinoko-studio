package chat

import (
	"strings"
	"unicode/utf8"

	"github.com/potproject/uchinoko-studio/api"
	"github.com/potproject/uchinoko-studio/data"
)

const chars = ".,?!;:—-)]} 。、？！；：」）］｝　\"'"

type ChatStream func(
	string, // apiKey
	[]data.CharacterConfigVoice, // voices
	bool, // multi
	string, // chatSystemPropmt
	string, // model
	[]data.ChatCompletionMessage, // messages
	string, // text
	*data.Image, // image
	chan api.ChunkMessage, // chunkMessage
) ([]data.ChatCompletionMessage, *data.Tokens, error)

func chatReceiver(
	charChannel chan rune,
	done chan error,
	multi bool,
	voices []data.CharacterConfigVoice,
	chunkMessage chan api.ChunkMessage,
	text string,
	image *data.Image,
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
		case err := <-done:
			if err != nil {
				return nil, err
			}
			chunkMessage <- api.TextChunkMessage{
				Text:  bufferText,
				Voice: voice,
			}
			if image != nil {
				text = "image"
			}
			return append(
				cm,
				data.ChatCompletionMessage{
					Role:    data.ChatCompletionMessageRoleUser,
					Content: text,
					Image:   image,
				},
				data.ChatCompletionMessage{
					Role:    data.ChatCompletionMessageRoleAssistant,
					Content: strings.Trim(allText, "\n"),
					Image:   nil,
				},
			), nil
		}
	}
}
