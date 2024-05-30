package router

import (
	"embed"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/potproject/uchinoko-studio/controller"
)

func Route(static embed.FS) *fiber.App {
	app := fiber.New()
	app.Use("/", filesystem.New(filesystem.Config{
		Root:       http.FS(static),
		PathPrefix: "static",
		Browse:     true,
	}))
	app.Static("/images/", "./images")
	app.Get("/adminopr/statall", controller.StatAll)
	app.Get("/adminopr/stat", controller.Stat)
	app.Get("/v1/chat/:id", controller.GetChat)
	app.Delete("/v1/chat/:id", controller.DeleteChat)
	app.Use("/v1/ws/talk", controller.WS)
	app.Get("/v1/ws/talk/:id/:characterId", controller.WSTalk())

	app.Get("/v1/config/general", controller.GetGeneralConfig)
	app.Post("/v1/config/general", controller.PostGeneralConfig)

	app.Get("/v1/config/characters", controller.GetCharacterConfigList)
	app.Get("/v1/config/character/init", controller.GetInitCharacterConfig)
	app.Post("/v1/config/character/:id", controller.PostCharacterConfig)
	app.Delete("/v1/config/character/:id", controller.DeleteCharacterConfig)

	return app
}
