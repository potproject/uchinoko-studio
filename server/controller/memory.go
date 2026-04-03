package controller

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/potproject/uchinoko-studio/data"
	"github.com/potproject/uchinoko-studio/memory"
)

func GetMemoryItems(c *fiber.Ctx) error {
	items, err := memory.ListMemoryItems(c.Params("ownerId"), c.Params("characterId"))
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusOK).JSON(items)
}

func PostMemoryItem(c *fiber.Ctx) error {
	item := new(data.MemoryItem)
	if err := c.BodyParser(item); err != nil {
		return err
	}
	item.OwnerID = strings.TrimSpace(c.Params("ownerId"))
	item.CharacterID = strings.TrimSpace(c.Params("characterId"))
	created, err := memory.CreateMemoryItem(*item)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusOK).JSON(created)
}

func PatchMemoryItem(c *fiber.Ctx) error {
	item := new(data.MemoryItem)
	if err := c.BodyParser(item); err != nil {
		return err
	}
	updated, err := memory.UpdateMemoryItem(c.Params("id"), *item)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusOK).JSON(updated)
}

func DeleteMemoryItem(c *fiber.Ctx) error {
	if err := memory.DeleteMemoryItem(c.Params("id")); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusOK)
}

func GetMemorySessionSummary(c *fiber.Ctx) error {
	summary, err := memory.GetSessionSummary(c.Params("ownerId"), c.Params("characterId"), c.Params("sessionId"))
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusOK).JSON(summary)
}

func RebuildCharacterMemory(c *fiber.Ctx) error {
	if err := memory.EnqueueRebuildCharacterMemory(c.Params("ownerId"), c.Params("characterId")); err != nil {
		return err
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "queued"})
}
