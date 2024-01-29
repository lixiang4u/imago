package utils

import (
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/lixiang4u/imago/models"
	"strings"
)

func TryParseUserId(ctx *fiber.Ctx) uint64 {
	var t = strings.TrimLeft(string(ctx.Request().Header.Peek("Authorization")), "Bearer")
	token, err := jwt.ParseWithClaims(strings.TrimSpace(t), jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(models.SECRET_KEY), nil
	})
	if err != nil {
		return 0
	}
	var claims = token.Claims.(jwt.MapClaims)
	return uint64(claims["id"].(float64))
}
