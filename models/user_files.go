package models

import "time"

type UserFiles struct {
	Id          uint64    `json:"id"`
	UserId      uint64    `json:"user_id"`
	ProxyId     uint64    `json:"proxy_id"`
	MetaId      string    `json:"meta_id"`
	OriginUrl   string    `json:"origin_url"`
	OriginFile  string    `json:"origin_file"`
	ConvertFile string    `json:"convert_file"`
	OriginSize  uint64    `json:"origin_size"`
	ConvertSize uint64    `json:"convert_size"`
	CreatedAt   time.Time `json:"created_at"`
}

func (UserFiles) TableName() string {
	return "user_files"
}

func GetOrCreateUserFiles(files UserFiles) (UserFiles, error) {
	var findUserFiles UserFiles
	if err := DB().Model(&findUserFiles).Where("user_id", files.UserId).Where("proxy_id", files.ProxyId).Where("meta_id", files.MetaId).Take(&findUserFiles).Error; err == nil {
		return findUserFiles, nil
	}
	if err := DB().Create(&files).Error; err != nil {
		return files, err
	}
	return files, nil
}
