package handlers

import (
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/lixiang4u/imago/models"
	"github.com/lixiang4u/imago/utils"
	"math/rand"
	"net/url"
	"strings"
	"time"
)

func Index(ctx *fiber.Ctx) error {
	return ctx.JSON(respSuccess(nil, "ok"))
}

func Debug(ctx *fiber.Ctx) error {
	return ctx.JSON(respSuccess(fiber.Map{
		"Hostname": ctx.Hostname(),
		"time":     time.Now().String(),
	}, ""))
}

func UserLogin(ctx *fiber.Ctx) error {
	type LoginRequest struct {
		Username string `json:"username" form:"username"`
		Password string `json:"password" form:"password"`
		Version  string `json:"version" json:"version"`
	}

	var loginRequest LoginRequest
	var host = ctx.Hostname()

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
	accessToken, err := utils.NewJwtAccessToken(u.Id, u.Nickname, host)
	if err != nil {
		return ctx.JSON(respErrorDebug("系统错误", err.Error()))
	}
	refreshToken, err := utils.NewJwtRefreshToken(u.Id, u.Nickname, host)
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

func checkOrigin(origin string) error {
	u, err := url.Parse(origin)
	if err != nil {
		return err
	}
	if len(u.Scheme) == 0 || len(u.Host) == 0 {
		return errors.New("源站地址格式错误")
	}
	if len(u.Path) > 0 || len(u.Query()) > 0 {
		return errors.New("源站地址格式错误")
	}
	return nil
}

func parseOrNewHost(ctx *fiber.Ctx) string {
	var tmpList = strings.Split(ctx.Hostname(), ".")
	if len(tmpList) <= 1 {
		return ""
	}
	var r = utils.HashString(fmt.Sprintf("%d,%d", time.Now().UnixNano(), rand.Int()))[:6]
	if len(tmpList) == 2 {
		tmpList = append([]string{r}, tmpList...)
	} else {
		tmpList[0] = r
	}
	return strings.Join(tmpList, ".")
}

func UserTokenRefresh(ctx *fiber.Ctx) error {
	claims := (ctx.Locals("user").(*jwt.Token)).Claims.(jwt.MapClaims)
	v, ok := claims["refresh"]
	if !ok {
		return ctx.JSON(respError("refresh_token错误", nil))
	}
	if len(v.(string)) == 0 {
		return ctx.JSON(respError("refresh_token异常", nil))
	}
	var id = claims["id"].(uint64)
	var name = claims["name"].(string)
	var iss = claims["iss"].(string)

	accessToken, err := utils.NewJwtAccessToken(id, name, iss)
	if err != nil {
		return ctx.JSON(respErrorDebug("系统错误", err.Error()))
	}
	refreshToken, err := utils.NewJwtRefreshToken(id, name, iss)
	if err != nil {
		return ctx.JSON(respErrorDebug("系统错误", err.Error()))
	}

	return ctx.JSON(respSuccess(fiber.Map{
		"user_id":       id,
		"nickname":      name,
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	}, "刷新成功"))
}

func CreateUserProxy(ctx *fiber.Ctx) error {
	claims := (ctx.Locals("user").(*jwt.Token)).Claims.(jwt.MapClaims)
	var userId = claims["id"].(uint64)

	type PostRequest struct {
		Title     string `json:"title" form:"title"`
		Origin    string `json:"origin" form:"origin"`
		Host      string `json:"host" form:"host"`
		Quality   int8   `json:"quality" form:"quality"`
		UserAgent string `json:"user_agent" form:"user_agent"`
		Cors      string `json:"cors" form:"cors"`
		Referer   string `json:"referer" form:"referer"`
	}

	var postRequest PostRequest
	if err := ctx.BodyParser(&postRequest); err != nil {
		return ctx.JSON(respError("参数错误", nil))
	}
	if err := checkOrigin(postRequest.Origin); err != nil {
		return ctx.JSON(respError(err.Error(), nil))
	}
	postRequest.Host = parseOrNewHost(ctx)
	if len(postRequest.Host) == 0 {
		return ctx.JSON(respError("生成代理域名失败，请重试", nil))
	}
	up, _ := models.GetHostUserProxy(postRequest.Host)
	if up.Id > 0 && up.UserId != userId {
		return ctx.JSON(respError("代理主机已存在", nil))
	}
	if models.GetUserProxyCount(userId) > 10 {
		return ctx.JSON(respError("代理数量超过限制", nil))
	}

	up = models.UserProxy{
		UserId:    userId,
		Title:     postRequest.Title,
		Origin:    postRequest.Origin,
		Host:      postRequest.Host,
		Quality:   postRequest.Quality,
		UserAgent: postRequest.UserAgent,
		Cors:      postRequest.Cors,
		Referer:   postRequest.Referer,
		Status:    models.PROXY_STATUS_OK,
		CreatedAt: time.Now(),
	}
	if err := models.DB().Create(&up).Error; err != nil {
		return ctx.JSON(respErrorDebug("创建失败", err.Error()))
	}

	return ctx.JSON(respSuccess(nil, "创建成功"))
}

func UpdateUserProxy(ctx *fiber.Ctx) error {
	var id = ctx.Params("id")
	claims := (ctx.Locals("user").(*jwt.Token)).Claims.(jwt.MapClaims)
	var userId = claims["id"].(uint64)

	type PostRequest struct {
		Title     string `json:"title" form:"title"`
		Origin    string `json:"origin" form:"origin"`
		Quality   int8   `json:"quality" form:"quality"`
		UserAgent string `json:"user_agent" form:"user_agent"`
		Cors      string `json:"cors" form:"cors"`
		Referer   string `json:"referer" form:"referer"`
		Status    int8   `json:"status" form:"status"`
	}

	var postRequest PostRequest
	if err := ctx.BodyParser(&postRequest); err != nil {
		return ctx.JSON(respError("参数错误", nil))
	}
	if err := checkOrigin(postRequest.Origin); err != nil {
		return ctx.JSON(respError(err.Error(), nil))
	}

	if err := models.DB().Model(&models.UserProxy{}).Where("id", id).Where("user_id", userId).Updates(map[string]interface{}{
		"title":      postRequest.Title,
		"origin":     postRequest.Origin,
		"quality":    postRequest.Quality,
		"user_agent": postRequest.UserAgent,
		"cors":       postRequest.Cors,
		"referer":    postRequest.Referer,
		"status":     postRequest.Status,
	}).Error; err != nil {
		return ctx.JSON(respErrorDebug("修改失败", err.Error()))
	}

	return ctx.JSON(respSuccess(nil, "创建成功"))
}

func DeleteUserProxy(ctx *fiber.Ctx) error {
	var id = ctx.Params("id")
	claims := (ctx.Locals("user").(*jwt.Token)).Claims.(jwt.MapClaims)
	var userId = claims["id"].(uint64)

	if err := models.DB().Where("id", id).Where("user_id", userId).Delete(&models.UserProxy{}).Error; err != nil {
		return ctx.JSON(respErrorDebug("删除失败", err.Error()))
	}

	return ctx.JSON(respSuccess(nil, "删除成功"))
}

func ListUserProxy(ctx *fiber.Ctx) error {
	var pager = utils.ParsePage(ctx)
	claims := (ctx.Locals("user").(*jwt.Token)).Claims.(jwt.MapClaims)
	var userId = claims["id"].(uint64)

	var ups []models.UserProxy
	engine := models.DB().Model(&ups).Where("user_id", userId)
	engine.Count(&pager.Total)
	engine.Limit(pager.Limit).Offset(pager.Offset).Order("id desc").Find(&ups)

	return ctx.JSON(respSuccessList(ups, pager, ""))
}

func ListUserProxyRequestLog(ctx *fiber.Ctx) error {
	var proxyId = ctx.Params("proxy_id")
	var pager = utils.ParsePage(ctx)
	claims := (ctx.Locals("user").(*jwt.Token)).Claims.(jwt.MapClaims)
	var userId = claims["id"].(uint64)

	type RespRequestLog struct {
		Id        uint64    `json:"id"`
		MetaId    string    `json:"meta_id"`
		OriginUrl string    `json:"origin_url"`
		Referer   string    `json:"referer"`
		Ip        string    `json:"ip"`
		IsCache   int8      `json:"is_cache"`
		IsExist   int8      `json:"is_exist"`
		CreatedAt time.Time `json:"created_at"`
	}

	var logs []RespRequestLog
	engine := models.DB().Model(&models.RequestLog{}).Where("user_id", userId).Where("proxy_id", proxyId)
	engine.Count(&pager.Total)
	engine.Limit(pager.Limit).Offset(pager.Offset).Order("id desc").Find(&logs)

	return ctx.JSON(respSuccessList(logs, pager, ""))
}
