package controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/potproject/uchinoko/db"
)

func StatAll(c *fiber.Ctx) error {
	j, e := db.ListAll()
	if e != nil {
		return e
	}
	return c.JSON(j)
}

func Stat(c *fiber.Ctx) error {
	j, e := db.ListAll()
	if e != nil {
		return e
	}
	keyOnly := []string{}
	for _, v := range j {
		keyOnly = append(keyOnly, v.Key)
	}
	return c.JSON(keyOnly)
}
