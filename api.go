package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/lixiang4u/imago/handlers"
	"github.com/lixiang4u/imago/models"
	"log"
)

func main() {
	app := fiber.New()
	app.Static("/upload", models.UploadRoot)
	app.Get("/ping", handlers.Ping)
	app.Post("/shrink", handlers.Shrink)
	log.Fatal(app.Listen(":8060"))
}
