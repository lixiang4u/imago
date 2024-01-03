package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/lixiang4u/imago/handlers"
	"log"
)

func main() {
	app := fiber.New()
	//app.Use(etag.New(etag.Config{Weak: true}))
	app.Get("/ping", handlers.Ping)
	app.Post("/shrink", handlers.Shrink)
	log.Fatal(app.Listen(":8060"))
}
