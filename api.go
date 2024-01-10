package main

import (
	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/lixiang4u/imago/handlers"
	"github.com/lixiang4u/imago/models"
	"log"
)

func main() {
	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "*",
		AllowMethods: "*",
	}))

	app.Static("/upload", models.UploadRoot)
	app.Get("/", handlers.Index)
	app.Get("/debug", handlers.Debug)
	app.Get("/ping", handlers.Ping)
	app.Post("/shrink", handlers.Shrink)

	app.Post("/user/login", handlers.UserLogin)
	app.Post("/user/register", handlers.UserRegister)

	// JWT Middleware
	app.Use(jwtware.New(jwtware.Config{
		SigningKey: jwtware.SigningKey{Key: []byte(models.SECRET_KEY)},
	}))

	app.Get("/user/status", handlers.UserTokenCheck)
	app.Get("/user/info", handlers.UserInfo)
	app.Get("/user/refresh-token", handlers.UserTokenRefresh)

	app.Post("/user/proxy", handlers.CreateUserProxy)
	app.Put("/user/proxy/:id", handlers.UpdateUserProxy)
	app.Delete("/user/proxy/:id", handlers.DeleteUserProxy)
	app.Get("/user/proxies", handlers.ListUserProxy)
	app.Get("/user/proxy/:proxy_id/logs", handlers.ListUserProxyRequestLog)

	log.Fatal(app.Listen(":8060"))
}
