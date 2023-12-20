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
	app.Get("/shrink", handlers.Shrink)
	app.Get("/*", handlers.Image)

	log.Fatal(app.Listen(":8020"))

}
