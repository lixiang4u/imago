package models

import (
	"errors"
	"fmt"
	"github.com/patrickmn/go-cache"
	"strings"
	"time"
)

var defaultAppConfig = AppConfig{
	UserId:     DEFAULT_UID,
	ProxyId:    0,
	OriginSite: LocalConfig.App.Remote,
	LocalPath:  LocalConfig.App.Local,
	ProxyHost:  "",
	Refresh:    0,
}

func GetHostUserConfig(host string) (AppConfig, error) {
	if len(host) == 0 {
		return defaultAppConfig, nil
	}
	v, ok := LocalCache.Get(host)
	if !ok {
		// 去数据库查找用户，并将id赋值给v
		// 查询不到也需要给个空数据，让他缓存下次不走数据库
		var d = cache.NoExpiration
		up, err := GetHostUserProxy(host)
		if err != nil || up.Status != PROXY_STATUS_OK {
			v = defaultAppConfig
			d = time.Duration(time.Minute) * 10
		} else {
			if len(up.UserAgent) == 0 {
				up.UserAgent = UserAgent
			}
			v = AppConfig{
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
		}
		SetLocalUserConfig(host, v, d)
	}
	if v.(AppConfig).ProxyHost != host {
		return defaultAppConfig, errors.New("源站未注册：" + host)
	}
	return v.(AppConfig), nil
}

func RequestMessage(host string) (int64, error) {
	return LocalCache.IncrementInt64(fmt.Sprintf("%s-request", host), 1)
}

func IncrementRequestOkCount(host string) (int64, error) {
	return LocalCache.IncrementInt64(fmt.Sprintf("%s-request-ok", host), 1)
}

func SetLocalUserConfig(host string, x interface{}, d time.Duration) {
	LocalCache.Set(host, x, d)
	LocalCache.Set(fmt.Sprintf("%s-request", host), int64(0), d)
	LocalCache.Set(fmt.Sprintf("%s-request-ok", host), int64(0), d)
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

func IncrementIpShrink(ip string) int64 {
	var u = fmt.Sprintf("global-ip-shrink-count")
	if _, ok := LocalCache.Get(ip); !ok {
		// 需要配合数据库
		LocalCache.Set(u, int64(1), time.Hour*24)
		return 1
	} else {
		n, _ := LocalCache.IncrementInt64(u, 1)
		return n
	}
}
