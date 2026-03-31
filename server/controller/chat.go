package controller

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/potproject/uchinoko-studio/db"
)

func resolveSessionID(c *fiber.Ctx) string {
	sessionID := strings.TrimSpace(c.Query("sessionId"))
	if sessionID == "" {
		return c.Params("id")
	}
	return sessionID
}

func GetChat(c *fiber.Ctx) error {
	//get id
	sessionID := resolveSessionID(c)
	characterId := c.Params("characterId")
	//get message
	d, init, err := db.GetChatMessage(sessionID, characterId)
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
	sessionID := resolveSessionID(c)
	characterId := c.Params("characterId")
	//delete message
	err := db.DeleteChatMessage(sessionID, characterId)
	if err != nil {
		return err
	}
	// No Content
	return c.SendStatus(fiber.StatusOK)
}

func GetChatSessions(c *fiber.Ctx) error {
	id := c.Params("id")
	characterId := c.Params("characterId")

	sessions, err := db.ListChatSessionsForOwner(id, characterId)
	if err != nil {
		return err
	}

	if err := c.JSON(sessions); err != nil {
		return err
	}
	c.Status(fiber.StatusOK)
	return nil
}
