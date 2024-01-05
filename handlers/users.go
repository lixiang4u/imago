package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/lixiang4u/imago/models"
	"github.com/lixiang4u/imago/utils"
)

func UserLogin(ctx *fiber.Ctx) error {
	type LoginRequest struct {
		Username string `json:"username" form:"username"`
		Password string `json:"password" form:"password"`
		Version  string `json:"version" json:"version"`
	}

	var loginRequest LoginRequest

	if err := ctx.BodyParser(&loginRequest); err != nil {
		return ctx.JSON(respError("参数错误", nil))
	}
	u, err := models.GetLoginUser(loginRequest.Username)
	if err != nil {
		return ctx.JSON(respError("用户名或者密码错误", nil))
	}
	if u.Password != utils.PasswordHash(loginRequest.Password) {
		return ctx.JSON(respError("用户名或者密码错误", nil))
	}
	accessToken, err := utils.NewJwtAccessToken(u.Id, u.Nickname)
	if err != nil {
		return ctx.JSON(respErrorDebug("系统错误", err.Error()))
	}
	refreshToken, err := utils.NewJwtRefreshToken(u.Id, u.Nickname)
	if err != nil {
		return ctx.JSON(respErrorDebug("系统错误", err.Error()))
	}

	return ctx.JSON(respSuccess(fiber.Map{
		"user_id":       u.Id,
		"nickname":      u.Nickname,
		"created_at":    u.CreatedAt,
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	}, "登录成功"))
}

func UserInfo(ctx *fiber.Ctx) error {
	user := ctx.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	name := claims["name"].(string)
	return ctx.SendString("Welcome " + name)
}
