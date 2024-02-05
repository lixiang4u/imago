package handlers

import (
	"github.com/gofiber/fiber/v2"
	"log"
)

// R 处理handler异常，防止崩溃
func R(h fiber.Handler) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		defer func() {
			if err := recover(); err != nil {
				log.Println("[recover]", err)
				_ = ctx.JSON(respError("系统错误"))
			}
		}()
		return h(ctx)
	}
}
