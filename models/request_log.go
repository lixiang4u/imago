package models

import "time"

type RequestLog struct {
	Id        uint64    `json:"id"`
	UserId    uint64    `json:"user_id"`
	ProxyId   uint64    `json:"proxy_id"`
	MetaId    string    `json:"meta_id"`
	OriginUrl string    `json:"origin_url"`
	Referer   string    `json:"referer"`
	Ip        string    `json:"ip"`
	IsCache   int8      `json:"is_cache"`
	IsExist   int8      `json:"is_exist"`
	CreatedAt time.Time `json:"created_at"`
}

func (RequestLog) TableName() string {
	return "request_log"
}
