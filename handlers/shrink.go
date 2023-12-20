package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/lixiang4u/imago/utils"
	"log"
	"time"
)

func Shrink(ctx *fiber.Ctx) error {
	var imgConfig = parseConfig(ctx)

	fh, err := ctx.FormFile("file")
	if err != nil {
		return ctx.JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	log.Println("[fh.Header]", utils.ToJsonString(fh.Header, true))
	log.Println("[imgConfig]", utils.ToJsonString(imgConfig, true))

	return ctx.JSON(fiber.Map{
		"status": "ok",
		"time":   time.Now().Format("2006-01-02 15:04:05"),
	})
}
