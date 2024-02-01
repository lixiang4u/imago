package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/lixiang4u/imago/utils"
	"time"
)

func Debug(ctx *fiber.Ctx) error {
	var userId = utils.TryParseUserId(ctx)
	return ctx.JSON(respSuccess(fiber.Map{
		"Hostname": ctx.Hostname(),
		"time":     time.Now().String(),
		"UA":       string(ctx.Request().Header.Peek("User-Agent")),
		"userId":   userId,
	}))
}
