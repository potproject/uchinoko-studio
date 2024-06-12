package controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/potproject/uchinoko-studio/db"
)

func GetChat(c *fiber.Ctx) error {
	//get id
	id := c.Params("id")
	characterId := c.Params("characterId")
	//get message
	d, init, err := db.GetChatMessage(id, characterId)
	if err != nil {
		return err
	}
	if init {
		// No Content
		return c.SendStatus(fiber.StatusNoContent)
	}
	// return message
	err = c.JSON(d)
	if err != nil {
		return err
	}
	c.Status(fiber.StatusOK)
	return nil

}

func DeleteChat(c *fiber.Ctx) error {
	//get id
	id := c.Params("id")
	characterId := c.Params("characterId")
	//delete message
	err := db.DeleteChatMessage(id, characterId)
	if err != nil {
		return err
	}
	// No Content
	return c.SendStatus(fiber.StatusOK)
}
