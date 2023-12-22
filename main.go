package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/lixiang4u/imago/handlers"
	"github.com/lixiang4u/imago/models"
	"log"
)

func main() {

	if models.LocalConfig.App.Prefetch {
		go func() { _ = handlers.Prefetch() }()
	}

	app := fiber.New()
	//app.Use(etag.New(etag.Config{Weak: true}))
	app.Get("/ping", handlers.Ping)
	app.Post("/shrink", handlers.Shrink)
	app.Get("/*", handlers.Image)

	log.Fatal(app.Listen(":8020"))

}
