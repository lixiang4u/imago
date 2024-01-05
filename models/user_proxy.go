package models

import "time"

// status 1.正常，0.未开启
const PROXY_STATUS_OK = 1
const PROXY_STATUS_ERR = 0

type UserProxy struct {
	Id        uint64    `json:"id"`
	UserId    uint64    `json:"user_id"`
	Title     string    `json:"title"`
	Origin    string    `json:"origin"`
	Host      string    `json:"host"`
	Quality   int8      `json:"quality"`
	UserAgent string    `json:"user_agent"`
	Cors      string    `json:"cors"`
	Referer   string    `json:"referer"`
	Status    int8      `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

func (UserProxy) TableName() string {
	return "user_proxy"
}

func GetHostUserProxy(host string) (userProxy UserProxy, err error) {
	if err := DB().Model(&userProxy).Where("host", host).Take(&userProxy).Error; err != nil {
		return userProxy, err
	}
	return userProxy, nil
}

func GetUserProxyCount(userId uint64) int64 {
	var count int64
	if err := DB().Model(&UserProxy{}).Where("user_id", userId).Count(&count).Error; err != nil {
		return count
	}
	return count
}
