package models

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
)

var db *gorm.DB

func init() {
	db = initDB()
}

func initDB() *gorm.DB {
	// 参考 https://github.com/go-sql-driver/mysql#dsn-data-source-name 获取详情
	db, err := gorm.Open(mysql.Open(LocalConfig.MySQL.Dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Println("[gorm.Open.Error]", err.Error())
		os.Exit(-1)
	}
	return db
}

func DB() *gorm.DB {
	if db != nil {
		return db
	}
	return initDB()
}
