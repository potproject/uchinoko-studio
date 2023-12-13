package controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/potproject/uchinoko/db"
)

func Stat(c *fiber.Ctx) error {
	j, e := db.ListAll()
	if e != nil {
		return e
	}
	return c.JSON(j)
}
