package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/potproject/uchinoko/controller"
)

func Route() *fiber.App {
	app := fiber.New()
	app.Static("/", "../public")
	app.Get("/adminopr/statall", controller.StatAll)
	app.Get("/adminopr/stat", controller.Stat)
	app.Get("/v1/chat/:id", controller.Chat)
	app.Use("/v1/ws/talk", controller.WS)
	app.Get("/v1/ws/talk/:id/:voiceType/:fileType", controller.WSTalk())

	return app
}
