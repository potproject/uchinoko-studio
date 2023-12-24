package controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/potproject/uchinoko-studio/db"
)

func GetChat(c *fiber.Ctx) error {
	//get id
	id := c.Params("id")
	//get message
	d, init, err := db.GetChatMessage(id)
	if err != nil {
		return err
	}
	if init {
		// No Content
		return c.SendStatus(fiber.StatusNoContent)
	}
	//return message
	return c.JSON(d)

}

func DeleteChat(c *fiber.Ctx) error {
	//get id
	id := c.Params("id")
	//delete message
	err := db.DeleteChatMessage(id)
	if err != nil {
		return err
	}
	// No Content
	return c.SendStatus(fiber.StatusOK)
}
