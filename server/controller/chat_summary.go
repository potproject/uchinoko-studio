package controller

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/potproject/uchinoko-studio/api"
	"github.com/potproject/uchinoko-studio/db"
	"github.com/potproject/uchinoko-studio/envgen"
)

const historyChatPropmt = `
以下の内容はAIとのチャットログです。
長くなっているため、これらを最大1000文字でまとめ、過去の履歴として要約してください。
また、システムプロンプト自体を含めないでください。

* テキストの形式について
- 「あなた:」 はユーザーの発言です。
- 「アシスタント:」 はアシスタントの発言です。
- 「システム:」 はシステムプロンプトです。

* 要約の形式について
- 要約は最大1000文字でお願いします。
- 要約は箇条書きで書いてください。
- 要約にはシステムプロンプトを含めないでください。
- 要約にはユーザーの発言とアシスタントの発言を含め、ニュアンスを損なわないようにしてください。
- 要約には過去の履歴を含め、過去の会話を思い出させるようにしてください。
- また、「これまでの要約」は既に要約された内容です。これらも含めて要約してください。
- また、1000文字に入りきれないようであれば省いても問題ありません。優先度が低い内容は省いてください。

* これまでの要約
{{history}}

`

func PostSummary(c *fiber.Ctx) error {
	//get character id
	characterID := c.Params("characterId")
	cc, err := db.GetCharacterConfig(characterID)
	if err != nil {
		return err
	}

	chatType := cc.Chat.Type
	chatModel := cc.Chat.Model

	//get id
	id := c.Params("id")
	//get message
	d, init, err := db.GetChatMessage(id)
	if err != nil {
		return err
	}
	if init {
		// No Content
		return c.SendStatus(fiber.StatusNoContent)
	}

	text := "システム:\n" + cc.Chat.SystemPrompt + "\n"
	for _, v := range d.Chat {
		if v.Role == "user" {
			text +=
				`あなた:\n` +
					v.Content +
					"\n"
		}
		if v.Role == "assistant" {
			text +=
				`アシスタント:\n` +
					v.Content +
					"\n"
		}
	}

	oldHistory := cc.History
	if oldHistory == "" {
		oldHistory = "これまでの要約はありません。"
	}

	// replaceする
	replacedHistoryChatPropmt := strings.Replace(historyChatPropmt, "{{history}}", oldHistory, -1)

	newHistory := ""
	if chatType == "openai" {
		newHistory, err = api.OpenAIChat(envgen.Get().OPENAI_API_KEY(), replacedHistoryChatPropmt, chatModel, text)
		if err != nil {
			return err
		}
	}
	if chatType == "anthropic" {
		newHistory, err = api.AnthropicChat(envgen.Get().ANTHROPIC_API_KEY(), replacedHistoryChatPropmt, chatModel, text)
		if err != nil {
			return err
		}
	}

	cc.History = newHistory

	err = db.PutCharacterConfig(characterID, cc)
	if err != nil {
		return err
	}

	//delete message
	err = db.DeleteChatMessage(id)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"history": newHistory,
	})
}
