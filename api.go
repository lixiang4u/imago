package main

import (
	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/lixiang4u/imago/handlers"
	"github.com/lixiang4u/imago/models"
	"log"
)

func main() {
	app := fiber.New()
	app.Static("/upload", models.UploadRoot)
	app.Use(jwtware.New(jwtware.Config{
		SigningKey: jwtware.SigningKey{Key: []byte(models.SECRET_KEY)},
	}))
	app.Get("/ping", handlers.Ping)
	app.Post("/shrink", handlers.Shrink)

	app.Post("/user/login", handlers.UserLogin)
	app.Get("/user/info", handlers.UserInfo)

	log.Fatal(app.Listen(":8060"))
}
