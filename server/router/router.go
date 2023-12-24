package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/potproject/uchinoko-studio/controller"
)

func Route() *fiber.App {
	app := fiber.New()
	app.Static("/", "../front/build")
	app.Get("/adminopr/statall", controller.StatAll)
	app.Get("/adminopr/stat", controller.Stat)
	app.Get("/v1/chat/:id", controller.GetChat)
	app.Delete("/v1/chat/:id", controller.DeleteChat)
	app.Use("/v1/ws/talk", controller.WS)
	app.Get("/v1/ws/talk/:id/:voiceType/:fileType", controller.WSTalk())

	return app
}
