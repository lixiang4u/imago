package handlers

import (
	"github.com/gofiber/fiber/v2"
	"time"
)

func Ping(ctx *fiber.Ctx) error {
	return ctx.JSON(respSuccess(fiber.Map{
		"time": time.Now().Format("2006-01-02 15:04:05"),
	}))
}
