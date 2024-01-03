package models

import "time"

type RequestStat struct {
	Id           uint64    `json:"id"`
	UserId       uint64    `json:"user_id"`
	ProxyId      uint64    `json:"proxy_id"`
	MetaId       string    `json:"meta_id"`
	OriginUrl    string    `json:"origin_url"`
	RequestCount uint64    `json:"request_count"`
	ResponseByte uint64    `json:"response_byte"`
	SavedByte    uint64    `json:"saved_byte"`
	CreatedAt    time.Time `json:"created_at"`
}

func (RequestStat) TableName() string {
	return "request_stat"
}

func GetOrCreateRequestStat(stat RequestStat) (RequestStat, error) {
	var findRequestStat RequestStat
	if err := DB().Model(&findRequestStat).Where("user_id", stat.UserId).Where("proxy_id", stat.ProxyId).Where("meta_id", stat.MetaId).Take(&findRequestStat).Error; err == nil {
		return findRequestStat, nil
	}
	if err := DB().Create(&stat).Error; err != nil {
		return stat, err
	}
	return stat, nil
}
