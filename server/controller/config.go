package controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/potproject/uchinoko/data"
	"github.com/potproject/uchinoko/db"
)

func postConfig(c *fiber.Ctx) error {
	id := c.Params("id")
	config := new(data.Config)
	if err := c.BodyParser(config); err != nil {
		return err
	}
	return db.PutConfig(id, *config)
}

func getConfig(c *fiber.Ctx) error {
	id := c.Params("id")
	config, err := db.GetConfig(id)
	if err != nil {
		return err
	}
	return c.JSON(config)
}
