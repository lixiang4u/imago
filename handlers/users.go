package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	utils2 "github.com/gofiber/fiber/v2/utils"
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

func UserRegister(ctx *fiber.Ctx) error {
	type RegisterRequest struct {
		Nickname string `json:"nickname" form:"nickname"`
		Email    string `json:"email" form:"email"`
		Password string `json:"password" form:"password"`
	}

	var registerRequest RegisterRequest
	if err := ctx.BodyParser(&registerRequest); err != nil {
		return ctx.JSON(respError("参数错误"))
	}
	if len(registerRequest.Password) < 6 {
		return ctx.JSON(respError("密码过于简单，请设置长度6位以上且包含特殊字符"))
	}
	if models.IncrementUserRegister() > 200 {
		return ctx.JSON(respError("注册火爆，请稍后", nil))
	}
	if _, err := models.GetLoginUser(registerRequest.Email); err == nil {
		return ctx.JSON(respError("用户已存在"))
	}
	var u = models.User{
		Nickname:  utils.FormatNickname(registerRequest.Email),
		Email:     registerRequest.Email,
		Password:  utils.PasswordHash(registerRequest.Password),
		ApiKey:    utils.HashString(fmt.Sprintf("%s%d", registerRequest.Password, time.Now().UnixNano())),
		CreatedAt: time.Now(),
	}
	if err := models.DB().Create(&u).Error; err != nil {
		return ctx.JSON(respErrorDebug("用户注册失败", err.Error()))
	}

	return ctx.JSON(respSuccess(nil, "注册成功"))
}

func UserLogin(ctx *fiber.Ctx) error {
	type LoginRequest struct {
		Email    string `json:"email" form:"email"`
		Password string `json:"password" form:"password"`
		Version  string `json:"version" json:"version"`
	}

	var loginRequest LoginRequest
	var host = ctx.Hostname()

	if err := ctx.BodyParser(&loginRequest); err != nil {
		return ctx.JSON(respError("参数错误"))
	}
	u, err := models.GetLoginUser(loginRequest.Email)
	if err != nil {
		return ctx.JSON(respError("用户名或者密码错误"))
	}
	if models.GetLoginErrorCount(u.Id) > 10 {
		return ctx.JSON(respError("登录异常，稍后再试"))
	}
	if u.Password != utils.PasswordHash(loginRequest.Password) {
		models.IncrementLoginError(u.Id)
		return ctx.JSON(respError("用户名或者密码错误"))
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

func UserTokenCheck(ctx *fiber.Ctx) error {
	user := ctx.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	return ctx.JSON(respSuccess(fiber.Map{
		"nickname":  claims["name"].(string),
		"timestamp": time.Now().Unix(),
	}, "ok"))
}

func UserInfo(ctx *fiber.Ctx) error {
	user := ctx.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	name := claims["name"].(string)
	return ctx.SendString("Welcome " + name)
}

func parseOrigin(origin string) (o string, err error) {
	u, err := url.Parse(origin)
	if err != nil {
		return origin, err
	}
	if len(u.Scheme) == 0 || len(u.Host) == 0 {
		return origin, errors.New("源站地址格式错误")
	}
	return fmt.Sprintf("%s://%s", u.Scheme, u.Host), nil
}

func parseOrNewHost(ctx *fiber.Ctx) string {
	var tmpList = strings.Split(ctx.Hostname(), ".")
	if len(tmpList) <= 1 {
		return ""
	}
	if strings.ToLower(tmpList[0]) == "www" {
		tmpList = tmpList[1:]
	}
	var r = utils.HashString(fmt.Sprintf("%d,%d", time.Now().UnixNano(), rand.Int()))[:6]
	tmpList = append([]string{r}, tmpList...)
	return strings.Join(tmpList, ".")
}

func UserTokenRefresh(ctx *fiber.Ctx) error {
	claims := (ctx.Locals("user").(*jwt.Token)).Claims.(jwt.MapClaims)
	v, ok := claims["refresh"]
	if !ok {
		return ctx.JSON(respError("refresh_token错误"))
	}
	if len(v.(string)) == 0 {
		return ctx.JSON(respError("refresh_token异常"))
	}
	var id = uint64(claims["id"].(float64))
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
	var userId = uint64(claims["id"].(float64))

	type PostRequest struct {
		Title     string `json:"title" form:"title"`
		Origin    string `json:"origin" form:"origin"`
		Host      string `json:"host" form:"host"`
		Quality   int8   `json:"quality" form:"quality"`
		UserAgent string `json:"user_agent" form:"user_agent"`
		Cors      string `json:"cors" form:"cors"`
		Referer   string `json:"referer" form:"referer"`
		Status    int8   `json:"status" form:"status"`
	}

	var err error
	var postRequest PostRequest
	if err := ctx.BodyParser(&postRequest); err != nil {
		return ctx.JSON(respError("参数错误"))
	}
	if postRequest.Origin, err = parseOrigin(postRequest.Origin); err != nil {
		return ctx.JSON(respError(err.Error()))
	}
	postRequest.Host = parseOrNewHost(ctx)
	if len(postRequest.Host) == 0 {
		return ctx.JSON(respError("生成代理域名失败，请重试"))
	}
	up, _ := models.GetHostUserProxy(postRequest.Host)
	if up.Id > 0 && up.UserId != userId {
		return ctx.JSON(respError("代理主机已存在"))
	}
	if models.GetUserProxyCount(userId) >= 10 {
		return ctx.JSON(respError("代理数量超过限制"))
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

	_ = notifyUpdateUserProxy(up)

	return ctx.JSON(respSuccess(nil, "创建成功"))
}

func UpdateUserProxy(ctx *fiber.Ctx) error {
	var id = ctx.Params("id")
	claims := (ctx.Locals("user").(*jwt.Token)).Claims.(jwt.MapClaims)
	var userId = uint64(claims["id"].(float64))

	type PostRequest struct {
		Title     string `json:"title" form:"title"`
		Origin    string `json:"origin" form:"origin"`
		Quality   int8   `json:"quality" form:"quality"`
		UserAgent string `json:"user_agent" form:"user_agent"`
		Cors      string `json:"cors" form:"cors"`
		Referer   string `json:"referer" form:"referer"`
		Status    int8   `json:"status" form:"status"`
	}

	var err error
	var postRequest PostRequest
	if err := ctx.BodyParser(&postRequest); err != nil {
		return ctx.JSON(respError("参数错误"))
	}
	if postRequest.Origin, err = parseOrigin(postRequest.Origin); err != nil {
		return ctx.JSON(respError(err.Error()))
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

	var up models.UserProxy
	models.DB().Model(&up).Where("id", id).Where("user_id", userId).Take(&up)
	_ = notifyUpdateUserProxy(up)

	return ctx.JSON(respSuccess(nil, "修改成功"))
}

func DeleteUserProxy(ctx *fiber.Ctx) error {
	var id = ctx.Params("id")
	claims := (ctx.Locals("user").(*jwt.Token)).Claims.(jwt.MapClaims)
	var userId = uint64(claims["id"].(float64))

	var up models.UserProxy
	if e := models.DB().Model(&up).Where("id", id).Where("user_id", userId).Take(&up).Error; e != nil {
		return ctx.JSON(respErrorDebug("数据不存在"))
	}
	if err := models.DB().Where("id", id).Where("user_id", userId).Delete(&models.UserProxy{}).Error; err != nil {
		return ctx.JSON(respErrorDebug("删除失败", err.Error()))
	}

	_ = notifyUpdateUserProxy(up)

	return ctx.JSON(respSuccess(nil, "删除成功"))
}

func ListUserProxy(ctx *fiber.Ctx) error {
	var pager = utils.ParsePage(ctx)
	claims := (ctx.Locals("user").(*jwt.Token)).Claims.(jwt.MapClaims)
	var userId = uint64(claims["id"].(float64))

	var ups []models.UserProxy
	engine := models.DB().Model(&ups).Where("user_id", userId)
	engine.Count(&pager.Total)
	engine.Limit(pager.Limit).Offset(pager.Offset).Order("id desc").Find(&ups)

	return ctx.JSON(respSuccessList(ups, pager))
}

func ListUserProxyRequestLog(ctx *fiber.Ctx) error {
	var proxyId = ctx.Params("proxy_id")
	var pager = utils.ParsePage(ctx)
	claims := (ctx.Locals("user").(*jwt.Token)).Claims.(jwt.MapClaims)
	var userId = uint64(claims["id"].(float64))

	type RespRequestLog struct {
		Id         uint64    `json:"id"`
		MetaId     string    `json:"meta_id"`
		RequestUrl string    `json:"request_url"`
		OriginUrl  string    `json:"origin_url"`
		Referer    string    `json:"referer"`
		Ip         string    `json:"ip"`
		IsCache    int8      `json:"is_cache"`
		IsExist    int8      `json:"is_exist"`
		CreatedAt  time.Time `json:"created_at"`
	}

	var logs []RespRequestLog
	engine := models.DB().Model(&models.RequestLog{}).Where("user_id", userId).Where("proxy_id", proxyId)
	engine.Count(&pager.Total)
	engine.Limit(pager.Limit).Offset(pager.Offset).Order("id desc").Find(&logs)

	return ctx.JSON(respSuccessList(logs, pager))
}

func ListUserProxyStat(ctx *fiber.Ctx) error {
	claims := (ctx.Locals("user").(*jwt.Token)).Claims.(jwt.MapClaims)
	var userId = uint64(claims["id"].(float64))

	type RespStat struct {
		ProxyCount    int64 `json:"proxy_count"`
		RequestCount  int64 `json:"request_count"`
		ResponseBytes int64 `json:"response_bytes"`
		SavedBytes    int64 `json:"saved_bytes"`
	}
	var respStat RespStat
	models.DB().Model(&models.RequestStat{}).Select(
		"SUM(request_count) AS request_count",
		"SUM(response_byte) AS response_bytes",
		"SUM(saved_byte) AS saved_bytes",
	).Where("user_id", userId).Take(&respStat)

	models.DB().Model(&models.UserProxy{}).Where("user_id", userId).Count(&respStat.ProxyCount)

	return ctx.JSON(respSuccess(respStat, "ok"))
}

func testCreateLog(userId uint64, proxyId uint64) {
	var start = time.Now().Truncate(time.Minute).Unix()
	var end = time.Now().Add(time.Hour * 24).Truncate(time.Minute).Unix()
	var log models.RequestStatRequestChart
	var t time.Time
	for {
		if start > end {
			break
		}

		t = time.Unix(start, 0)
		if e := models.DB().Model(&log).Where("user_id", userId).Where("proxy_id", proxyId).Where("created_at", time.Unix(start, 0)).Take(&log); e != nil {
			// 五
			models.DB().Create(&models.RequestStatRequestChart{
				UserId:       userId,
				ProxyId:      proxyId,
				RequestCount: uint64(rand.Intn(9999)),
				ResponseByte: uint64(rand.Intn(999999)),
				SavedByte:    uint64(rand.Intn(999999)),
				CreatedAt:    t,
			})
		} else {

		}

		start += 60
	}

}

func ListUserProxyProxyRequestStat(ctx *fiber.Ctx) error {
	var proxyId = utils.StringToUint64(ctx.Params("proxy_id"))
	claims := (ctx.Locals("user").(*jwt.Token)).Claims.(jwt.MapClaims)
	var userId = uint64(claims["id"].(float64))
	date, dateErr := time.Parse("2006-01-02", ctx.Query("date"))
	if dateErr != nil {
		date = time.Now().Add(-time.Hour * 12)
	}

	var logs []models.RequestStatRequestChart
	engine := models.DB().Model(&logs).Where("user_id", userId)
	if proxyId > 0 {
		engine.Where("proxy_id", proxyId)
	}
	engine.Where("created_at >= ?", utils2.ToString(date))
	engine.Where("created_at < ?", utils2.ToString(date.AddDate(0, 0, 1)))
	engine.Order("created_at ASC").Limit(5000).Find(&logs)

	if len(logs) == 0 {
		return ctx.JSON(respSuccess(nil, "没有数据"))
	}

	// 60 *24
	var start = logs[0].CreatedAt.Truncate(time.Minute).Unix()
	var end = logs[len(logs)-1].CreatedAt.Unix()
	if end-start > 86400 {
		end = start + 86400
	}

	var logMap = make(map[int64]models.RequestStatRequestChart)
	for _, log := range logs {
		v, ok := logMap[log.CreatedAt.Unix()]
		if !ok {
			logMap[log.CreatedAt.Unix()] = log
		} else {
			v.RequestCount += log.RequestCount
			v.ResponseByte += log.ResponseByte
			v.SavedByte += log.SavedByte
			logMap[log.CreatedAt.Unix()] = v
		}
	}

	type RespLog struct {
		T        string `json:"t"`
		Count    uint64 `json:"count"`
		RespByte uint64 `json:"resp_byte"`
		SaveByte uint64 `json:"save_byte"`
	}
	var respLogs []RespLog

	for {
		if start > end {
			break
		}
		t := time.Unix(start, 0)
		v, ok := logMap[start]
		if !ok {
			respLogs = append(respLogs, RespLog{
				T:        t.Format("15:04"),
				Count:    0,
				RespByte: 0,
				SaveByte: 0,
			})
		} else {
			respLogs = append(respLogs, RespLog{
				T:        t.Format("15:04"),
				Count:    v.RequestCount,
				RespByte: v.RequestCount,
				SaveByte: v.SavedByte,
			})
		}
		start += 60
	}

	return ctx.JSON(respSuccess(respLogs, "ok"))
}

func notifyUpdateUserProxy(proxy models.UserProxy) error {
	buf, err := json.Marshal(models.AdminCommand{
		Command: models.NsqCmd0x0010,
		Body:    utils.ToJsonBytes(proxy),
	})
	if err != nil {
		return err
	}
	err = models.NsqProducer(models.TopicAdminCommand, buf)
	if err != nil {
		return err
	}
	return nil
}
