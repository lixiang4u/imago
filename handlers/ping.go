package handlers

import (
	"github.com/gofiber/fiber/v2"
	"time"
)

func Ping(ctx *fiber.Ctx) error {
	return ctx.JSON(fiber.Map{"status": "ok", "time": time.Now().Format("2006-01-02 15:04:05")})
}
