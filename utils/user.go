package utils

import (
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func TryParseUserId(ctx *fiber.Ctx) uint64 {
	var u = ctx.Locals("user")
	if u == nil {
		return 0
	}
	var claims = u.(*jwt.Token).Claims.(jwt.MapClaims)
	return uint64(claims["id"].(float64))
}
