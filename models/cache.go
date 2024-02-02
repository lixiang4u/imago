package models

import (
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/patrickmn/go-cache"
	"strings"
	"time"
)

func GetHostCacheKey(host string, userId ...uint64) string {
	if len(userId) > 0 && userId[0] > 0 {
		return fmt.Sprintf("%d@%s", userId[0], host)
	}
	return host
}

func GetHostUserConfig(ctx *fiber.Ctx, userId ...uint64) (AppConfig, error) {
	var host = string(ctx.Request().Host())
	host = GetHostCacheKey(host, userId...)
	v, ok := LocalCache.Get(host)
	if ok {
		return v.(AppConfig), nil
	}

	//  这里走的是代理，所以无法判断出正确的前端请求protocol
	var newOrigin = "https://" + host

	up, err := GetHostUserProxy(host, userId...)
	if err != nil {
		// 指定用户，但是用户ID为真（普通用户）
		if len(userId) > 0 && userId[0] > 0 {
			up = CreateDefaultUserProxy(userId[0], newOrigin, host) // 创建用户默认代理
		}
		// 指定用户，但是用户ID为0（Guest用户）
		if len(userId) > 0 && userId[0] <= 0 {
			return AppConfig{UserId: 0, ProxyId: 0, OriginSite: newOrigin, ProxyHost: host}, nil
		}
	}
	if up.Id <= 0 {
		return AppConfig{}, errors.New(fmt.Sprintf("代理不存在：%s", host))
	}
	if len(up.UserAgent) == 0 {
		up.UserAgent = UserAgent
	}
	var tmpConfig = AppConfig{
		UserId:     up.UserId,
		ProxyId:    up.Id,
		OriginSite: strings.TrimSpace(up.Origin), //https://abc.imago-service.xyz
		UserAgent:  strings.TrimSpace(up.UserAgent),
		Cors:       strings.TrimSpace(up.Cors),
		ProxyHost:  strings.TrimSpace(up.Host), //abc.imago-service.xyz
		LocalPath:  "",
		Refresh:    0,
		Debug:      true,
	}
	SetLocalUserConfig(&tmpConfig, tmpConfig, cache.NoExpiration)
	return tmpConfig, nil
}

func IncrementRequestCount(appConfig *AppConfig) (int64, error) {
	return LocalCache.IncrementInt64(fmt.Sprintf("%s-request", GetHostCacheKey(appConfig.ProxyHost, appConfig.UserId)), 1)
}

func IncrementRequestOkCount(appConfig *AppConfig) (int64, error) {
	return LocalCache.IncrementInt64(fmt.Sprintf("%s-request-ok", GetHostCacheKey(appConfig.ProxyHost, appConfig.UserId)), 1)
}

func SetLocalUserConfig(appConfig *AppConfig, x interface{}, d time.Duration) {
	var cacheKey = GetHostCacheKey(appConfig.ProxyHost, appConfig.UserId)
	LocalCache.Set(cacheKey, x, d)
	LocalCache.Set(fmt.Sprintf("%s-request", cacheKey), int64(0), d)
	LocalCache.Set(fmt.Sprintf("%s-request-ok", cacheKey), int64(0), d)
}

func IncrementLoginError(userId uint64) int64 {
	var u = fmt.Sprintf("login-error-%d", userId)
	if _, ok := LocalCache.Get(u); !ok {
		LocalCache.Set(u, int64(1), time.Hour)
		return 1
	} else {
		n, _ := LocalCache.IncrementInt64(u, 1)
		return n
	}
}

func GetLoginErrorCount(userId uint64) int64 {
	var u = fmt.Sprintf("login-error-%d", userId)
	if v, ok := LocalCache.Get(u); ok {
		return v.(int64)
	}
	return 0
}

func IncrementUserRegister() int64 {
	var u = fmt.Sprintf("global-register-count")
	if _, ok := LocalCache.Get(u); !ok {
		LocalCache.Set(u, int64(1), time.Hour)
		return 1
	} else {
		n, _ := LocalCache.IncrementInt64(u, 1)
		return n
	}
}

func IncrementWebShrinkCount(cacheKey string) int64 {
	var u = fmt.Sprintf("global-user-shrink-count")
	if _, ok := LocalCache.Get(cacheKey); !ok {
		// 需要配合数据库
		LocalCache.Set(u, int64(1), time.Hour*24)
		return 1
	} else {
		n, _ := LocalCache.IncrementInt64(u, 1)
		return n
	}
}
