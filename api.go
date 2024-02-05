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
		// https://developer.mozilla.org/zh-CN/docs/Glossary/Simple_response_header
		ExposeHeaders: "Content-Disposition, X-zip_name",
	}))

	app.Get("/", handlers.R(handlers.Index))
	app.Get("/debug", handlers.R(handlers.Debug))
	app.Get("/ping", handlers.R(handlers.Ping))
	app.Post("/upload", handlers.R(handlers.Upload))
	app.Post("/process", handlers.R(handlers.Process))
	app.Get("/file/*", handlers.R(handlers.Download))
	app.Post("/archive/zip", handlers.R(handlers.Archive))

	app.Post("/user/login", handlers.R(handlers.UserLogin))
	app.Post("/user/register", handlers.R(handlers.UserRegister))

	// JWT Middleware
	app.Use(jwtware.New(jwtware.Config{
		SigningKey: jwtware.SigningKey{Key: []byte(models.SECRET_KEY)},
	}))

	app.Get("/user/status", handlers.R(handlers.UserTokenCheck))
	app.Get("/user/info", handlers.R(handlers.UserInfo))
	app.Get("/user/refresh-token", handlers.R(handlers.UserTokenRefresh))

	app.Post("/user/proxy", handlers.R(handlers.CreateUserProxy))
	app.Put("/user/proxy/:id", handlers.R(handlers.UpdateUserProxy))
	app.Delete("/user/proxy/:id", handlers.R(handlers.DeleteUserProxy))
	app.Get("/user/proxies", handlers.R(handlers.ListUserProxy))
	app.Get("/user/proxy/:proxy_id/logs", handlers.R(handlers.ListUserProxyRequestLog))
	app.Get("/user/proxy/stat", handlers.R(handlers.ListUserProxyStat))
	app.Get("/proxy/:proxy_id/request/stat", handlers.R(handlers.ListUserProxyProxyRequestStat))

	log.Fatal(app.Listen(":8060"))
}
