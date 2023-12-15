package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/lixiang4u/imago/handlers"
	"log"
)

func main() {

	app := fiber.New()

	app.Get("/ping", func(ctx *fiber.Ctx) error {
		return ctx.JSON(fiber.Map{"status": "ok"})
	})

	app.Get("/*", handlers.Image)

	log.Fatal(app.Listen(":8020"))

}
