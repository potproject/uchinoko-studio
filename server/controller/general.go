package controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/potproject/uchinoko-studio/data"
	"github.com/potproject/uchinoko-studio/db"
)

func PostGeneralConfig(c *fiber.Ctx) error {
	config := new(data.GeneralConfig)
	if err := c.BodyParser(config); err != nil {
		return err
	}
	return db.PutGeneralConfig(*config)
}

func GetGeneralConfig(c *fiber.Ctx) error {
	config, err := db.GetGeneralConfig()
	if err != nil {
		return err
	}
	err = c.JSON(config)
	if err != nil {
		return err
	}
	c.Status(fiber.StatusOK)
	return nil
}
