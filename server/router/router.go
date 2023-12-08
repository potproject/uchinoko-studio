package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/potproject/uchinoko/controller"
)

func Route() *fiber.App {
	app := fiber.New()
	app.Static("/", "../public")
	app.Use("/v1/ws/talk", controller.WS)
	app.Get("/v1/ws/talk/:id/:fileType", controller.WSTalk())

	return app
}
