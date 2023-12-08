package controller

import (
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

func WS(c *fiber.Ctx) error {
	// IsWebSocketUpgrade returns true if the client
	// requested upgrade to the WebSocket protocol.
	if websocket.IsWebSocketUpgrade(c) {
		c.Locals("allowed", true)
		return c.Next()
	}
	return fiber.ErrUpgradeRequired
}
