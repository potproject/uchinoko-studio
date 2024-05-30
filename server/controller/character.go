package controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/potproject/uchinoko-studio/data"
	"github.com/potproject/uchinoko-studio/db"
)

func PostCharacterConfig(c *fiber.Ctx) error {
	id := c.Params("id")
	config := new(data.CharacterConfig)
	if err := c.BodyParser(config); err != nil {
		return err
	}
	return db.PutCharacterConfig(id, *config)
}

func DeleteCharacterConfig(c *fiber.Ctx) error {
	id := c.Params("id")
	return db.DeleteCharacterConfig(id)
}

func GetCharacterConfigList(c *fiber.Ctx) error {
	config, err := db.GetCharacterConfigList()
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

func GetInitCharacterConfig(c *fiber.Ctx) error {
	config := db.CharacterInitConfig()
	return c.JSON(config)
}
