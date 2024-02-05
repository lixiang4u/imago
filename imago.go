package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/lixiang4u/imago/handlers"
	"github.com/lixiang4u/imago/models"
	"log"
)

func main() {

	var h = handlers.NsqConsumeHandler{}
	if err := h.HandleMessage(); err != nil {
		log.Println("[nsq.HandleMessage.Error]", err.Error())
		return
	}
	defer h.NsqStop()

	if models.LocalConfig.App.Prefetch {
		go func() { _ = handlers.Prefetch() }()
	}

	app := fiber.New()
	//app.Use(etag.New(etag.Config{Weak: true}))
	app.Get("/*", handlers.R(handlers.Image))

	log.Fatal(app.Listen(":8020"))

}
