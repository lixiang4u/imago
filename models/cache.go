package models

import (
	"errors"
	"fmt"
	"github.com/patrickmn/go-cache"
)

var defaultAppConfig = AppConfig{
	UserId:     DEFAULT_UID,
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
		v = AppConfig{
			UserId:     "100002",
			OriginSite: "https://abc.imago-service.xyz",
			LocalPath:  "",
			ProxyHost:  "abc.imago-service.xyz",
			Refresh:    0,
		}
		LocalCache.Set(host, v, cache.NoExpiration)
		LocalCache.Set(fmt.Sprintf("%s-request", host), int64(0), cache.NoExpiration)
		LocalCache.Set(fmt.Sprintf("%s-request-ok", host), int64(0), cache.NoExpiration)
	}
	if v.(AppConfig).ProxyHost != host {
		return defaultAppConfig, errors.New("源站未注册：" + host)
	}
	return v.(AppConfig), nil
}

func IncrementRequestCount(host string) (int64, error) {
	//
	return LocalCache.IncrementInt64(fmt.Sprintf("%s-request", host), 1)
}

func IncrementRequestOkCount(host string) (int64, error) {
	return LocalCache.IncrementInt64(fmt.Sprintf("%s-request-ok", host), 1)
}
