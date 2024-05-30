package router

import (
	"embed"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/potproject/uchinoko-studio/controller"
	"github.com/potproject/uchinoko-studio/envgen"
)

func Route(static embed.FS) *fiber.App {
	app := fiber.New()
	// Frontend
	app.Use("/", filesystem.New(filesystem.Config{
		Root:       http.FS(static),
		PathPrefix: "static",
		Browse:     true,
	}))
	app.Static("/images/", "./images")

	// API
	if envgen.Get().DEBUG() && !envgen.Get().READ_ONLY() {
		debugRoutes(app)
	}

	if !envgen.Get().READ_ONLY() {
		managerRoutes(app)
	}

	userRoutes(app)

	return app
}

func debugRoutes(app *fiber.App) {
	app.Get("/v1/admin/statall", controller.StatAll)
	app.Get("/v1/admin/stat", controller.Stat)
}

func userRoutes(app *fiber.App) {
	app.Get("/v1/chat/:id", controller.GetChat)
	app.Delete("/v1/chat/:id", controller.DeleteChat)
	app.Use("/v1/ws/talk", controller.WS)
	app.Get("/v1/ws/talk/:id/:characterId", controller.WSTalk())

	app.Get("/v1/config/general", controller.GetGeneralConfig)

	app.Get("/v1/config/characters", controller.GetCharacterConfigList)
	app.Get("/v1/config/character/init", controller.GetInitCharacterConfig)
}

func managerRoutes(app *fiber.App) {
	app.Post("/v1/config/general", controller.PostGeneralConfig)

	app.Post("/v1/config/character/:id", controller.PostCharacterConfig)
	app.Delete("/v1/config/character/:id", controller.DeleteCharacterConfig)
}
