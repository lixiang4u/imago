package models

import "time"

type RequestStatRequestChart struct {
	Id           uint64    `json:"id"`
	UserId       uint64    `json:"user_id"`
	ProxyId      uint64    `json:"proxy_id"`
	RequestCount uint64    `json:"request_count"`
	ResponseByte uint64    `json:"response_byte"`
	SavedByte    uint64    `json:"saved_byte"`
	CreatedAt    time.Time `json:"created_at"`
}

func (RequestStatRequestChart) TableName() string {
	return "request_stat_request_chart"
}

func GetOrCreateRequestStatRequestChart(stat RequestStatRequestChart) (RequestStatRequestChart, error) {
	var findRequestStatRequestChart RequestStatRequestChart
	if err := DB().Model(&findRequestStatRequestChart).Where("user_id", stat.UserId).Where("proxy_id", stat.ProxyId).Where("created_at", stat.CreatedAt).Take(&findRequestStatRequestChart).Error; err == nil {
		return findRequestStatRequestChart, nil
	}
	if err := DB().Create(&stat).Error; err != nil {
		return stat, err
	}
	return stat, nil
}
