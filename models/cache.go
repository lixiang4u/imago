package models

import (
	"errors"
	"fmt"
	"github.com/patrickmn/go-cache"
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
			v = AppConfig{
				UserId:     up.UserId,
				ProxyId:    up.Id,
				OriginSite: up.Origin, //https://abc.imago-service.xyz
				LocalPath:  "",
				ProxyHost:  up.Host, //abc.imago-service.xyz
				Refresh:    0,
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
