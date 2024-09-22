package controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/potproject/uchinoko-studio/db"
	"github.com/potproject/uchinoko-studio/envgen"
)

type EnvConfigResponse struct {
	OPENAI_SPEECH_TO_TEXT_API_KEY bool `json:"OPENAI_SPEECH_TO_TEXT_API_KEY"`
	GOOGLE_SPEECH_TO_TEXT_API_KEY bool `json:"GOOGLE_SPEECH_TO_TEXT_API_KEY"`
	VOSK_SERVER_ENDPOINT          bool `json:"VOSK_SERVER_ENDPOINT"`
	OPENAI_API_KEY                bool `json:"OPENAI_API_KEY"`
	ANTHROPIC_API_KEY             bool `json:"ANTHROPIC_API_KEY"`
	COHERE_API_KEY                bool `json:"COHERE_API_KEY"`
	GEMINI_API_KEY                bool `json:"GEMINI_API_KEY"`
	OPENAI_LOCAL_API_KEY          bool `json:"OPENAI_LOCAL_API_KEY"`
	OPENAI_LOCAL_API_ENDPOINT     bool `json:"OPENAI_LOCAL_API_ENDPOINT"`
	VOICEVOX_ENDPOINT             bool `json:"VOICEVOX_ENDPOINT"`
	BERTVITS2_ENDPOINT            bool `json:"BERTVITS2_ENDPOINT"`
	STYLEBERTVIT2_ENDPOINT        bool `json:"STYLEBERTVIT2_ENDPOINT"`
	GOOGLE_TEXT_TO_SPEECH_API_KEY bool `json:"GOOGLE_TEXT_TO_SPEECH_API_KEY"`
	OPENAI_SPEECH_API_KEY         bool `json:"OPENAI_SPEECH_API_KEY"`
}

// GetEnvConfig 環境変数の現在の設定状態を取得する
func GetEnvConfig(c *fiber.Ctx) error {
	config, err := db.GetEnvConfig()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "環境変数の取得に失敗しました",
		})
	}

	// envgenパッケージを使用して、環境変数が設定されているかを確認
	env := envgen.Get()

	response := EnvConfigResponse{
		OPENAI_SPEECH_TO_TEXT_API_KEY: env.OPENAI_SPEECH_TO_TEXT_API_KEY() != "" || config.OPENAI_SPEECH_TO_TEXT_API_KEY != "",
		GOOGLE_SPEECH_TO_TEXT_API_KEY: env.GOOGLE_SPEECH_TO_TEXT_API_KEY() != "" || config.GOOGLE_SPEECH_TO_TEXT_API_KEY != "",
		VOSK_SERVER_ENDPOINT:          env.VOSK_SERVER_ENDPOINT() != "" || config.VOSK_SERVER_ENDPOINT != "",
		OPENAI_API_KEY:                env.OPENAI_API_KEY() != "" || config.OPENAI_API_KEY != "",
		ANTHROPIC_API_KEY:             env.ANTHROPIC_API_KEY() != "" || config.ANTHROPIC_API_KEY != "",
		COHERE_API_KEY:                env.COHERE_API_KEY() != "" || config.COHERE_API_KEY != "",
		GEMINI_API_KEY:                env.GEMINI_API_KEY() != "" || config.GEMINI_API_KEY != "",
		OPENAI_LOCAL_API_KEY:          env.OPENAI_LOCAL_API_KEY() != "" || config.OPENAI_LOCAL_API_KEY != "",
		OPENAI_LOCAL_API_ENDPOINT:     env.OPENAI_LOCAL_API_ENDPOINT() != "" || config.OPENAI_LOCAL_API_ENDPOINT != "",
		VOICEVOX_ENDPOINT:             env.VOICEVOX_ENDPOINT() != "" || config.VOICEVOX_ENDPOINT != "",
		BERTVITS2_ENDPOINT:            env.BERTVITS2_ENDPOINT() != "" || config.BERTVITS2_ENDPOINT != "",
		STYLEBERTVIT2_ENDPOINT:        env.STYLEBERTVIT2_ENDPOINT() != "" || config.STYLEBERTVIT2_ENDPOINT != "",
		GOOGLE_TEXT_TO_SPEECH_API_KEY: env.GOOGLE_TEXT_TO_SPEECH_API_KEY() != "" || config.GOOGLE_TEXT_TO_SPEECH_API_KEY != "",
		OPENAI_SPEECH_API_KEY:         env.OPENAI_SPEECH_API_KEY() != "" || config.OPENAI_SPEECH_API_KEY != "",
	}

	err = c.JSON(response)
	if err != nil {
		return err
	}
	c.Status(fiber.StatusOK)
	return nil
}

// PostEnvConfig 環境変数の設定を更新する
func PostEnvConfig(c *fiber.Ctx) error {
	if envgen.Get().READ_ONLY() {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "設定の更新は許可されていません",
		})
	}

	var data map[string]string
	if err := c.BodyParser(&data); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "無効なリクエストボディです",
		})
	}

	// 更新可能な環境変数のリスト
	validKeys := map[string]bool{
		"OPENAI_SPEECH_TO_TEXT_API_KEY": true,
		"GOOGLE_SPEECH_TO_TEXT_API_KEY": true,
		"VOSK_SERVER_ENDPOINT":          true,
		"OPENAI_API_KEY":                true,
		"ANTHROPIC_API_KEY":             true,
		"COHERE_API_KEY":                true,
		"GEMINI_API_KEY":                true,
		"OPENAI_LOCAL_API_KEY":          true,
		"OPENAI_LOCAL_API_ENDPOINT":     true,
		"VOICEVOX_ENDPOINT":             true,
		"BERTVITS2_ENDPOINT":            true,
		"STYLEBERTVIT2_ENDPOINT":        true,
		"GOOGLE_TEXT_TO_SPEECH_API_KEY": true,
		"OPENAI_SPEECH_API_KEY":         true,
	}

	// 現在の環境変数設定を取得
	config, err := db.GetEnvConfig()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "環境変数の取得に失敗しました",
		})
	}

	// リクエストのデータで環境変数を更新
	changed := false
	for key, value := range data {
		if validKeys[key] {
			switch key {
			case "OPENAI_SPEECH_TO_TEXT_API_KEY":
				if config.OPENAI_SPEECH_TO_TEXT_API_KEY != value {
					config.OPENAI_SPEECH_TO_TEXT_API_KEY = value
					envgen.Set().OPENAI_SPEECH_TO_TEXT_API_KEY(value)
					changed = true
				}
			case "GOOGLE_SPEECH_TO_TEXT_API_KEY":
				if config.GOOGLE_SPEECH_TO_TEXT_API_KEY != value {
					config.GOOGLE_SPEECH_TO_TEXT_API_KEY = value
					envgen.Set().GOOGLE_SPEECH_TO_TEXT_API_KEY(value)
					changed = true
				}
			case "VOSK_SERVER_ENDPOINT":
				if config.VOSK_SERVER_ENDPOINT != value {
					config.VOSK_SERVER_ENDPOINT = value
					envgen.Set().VOSK_SERVER_ENDPOINT(value)
					changed = true
				}
			case "OPENAI_API_KEY":
				if config.OPENAI_API_KEY != value {
					config.OPENAI_API_KEY = value
					envgen.Set().OPENAI_API_KEY(value)
					changed = true
				}
			case "ANTHROPIC_API_KEY":
				if config.ANTHROPIC_API_KEY != value {
					config.ANTHROPIC_API_KEY = value
					envgen.Set().ANTHROPIC_API_KEY(value)
					changed = true
				}
			case "COHERE_API_KEY":
				if config.COHERE_API_KEY != value {
					config.COHERE_API_KEY = value
					envgen.Set().COHERE_API_KEY(value)
					changed = true
				}
			case "GEMINI_API_KEY":
				if config.GEMINI_API_KEY != value {
					config.GEMINI_API_KEY = value
					envgen.Set().GEMINI_API_KEY(value)
					changed = true
				}
			case "OPENAI_LOCAL_API_KEY":
				if config.OPENAI_LOCAL_API_KEY != value {
					config.OPENAI_LOCAL_API_KEY = value
					envgen.Set().OPENAI_LOCAL_API_KEY(value)
					changed = true
				}
			case "OPENAI_LOCAL_API_ENDPOINT":
				if config.OPENAI_LOCAL_API_ENDPOINT != value {
					config.OPENAI_LOCAL_API_ENDPOINT = value
					envgen.Set().OPENAI_LOCAL_API_ENDPOINT(value)
					changed = true
				}
			case "VOICEVOX_ENDPOINT":
				if config.VOICEVOX_ENDPOINT != value {
					config.VOICEVOX_ENDPOINT = value
					envgen.Set().VOICEVOX_ENDPOINT(value)
					changed = true
				}
			case "BERTVITS2_ENDPOINT":
				if config.BERTVITS2_ENDPOINT != value {
					config.BERTVITS2_ENDPOINT = value
					envgen.Set().BERTVITS2_ENDPOINT(value)
					changed = true
				}
			case "STYLEBERTVIT2_ENDPOINT":
				if config.STYLEBERTVIT2_ENDPOINT != value {
					config.STYLEBERTVIT2_ENDPOINT = value
					envgen.Set().STYLEBERTVIT2_ENDPOINT(value)
					changed = true
				}
			case "GOOGLE_TEXT_TO_SPEECH_API_KEY":
				if config.GOOGLE_TEXT_TO_SPEECH_API_KEY != value {
					config.GOOGLE_TEXT_TO_SPEECH_API_KEY = value
					envgen.Set().GOOGLE_TEXT_TO_SPEECH_API_KEY(value)
					changed = true
				}
			case "OPENAI_SPEECH_API_KEY":
				if config.OPENAI_SPEECH_API_KEY != value {
					config.OPENAI_SPEECH_API_KEY = value
					envgen.Set().OPENAI_SPEECH_API_KEY(value)
					changed = true
				}
			}
		}
	}

	if changed {
		// データベースに保存
		err := db.PutEnvConfig(config)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "環境変数の保存に失敗しました",
			})
		}
	}

	return c.JSON(fiber.Map{
		"message": "環境変数が更新されました",
	})
}
