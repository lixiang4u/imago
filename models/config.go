package models

import (
	"github.com/spf13/viper"
	"log"
	"os"
)

var (
	ConfigRemote string
	ConfigLocal  string
)

func init() {
	ConfigLocal = UploadRoot

	if _, err := os.Stat("config.toml"); err != nil {
		log.Println("file config.toml, ", err.Error())
		return
	}
	viper.SetConfigFile("config.toml")
	err := viper.ReadInConfig()
	if err != nil {
		log.Println("read config.toml, ", err.Error())
		return
	}

	ConfigRemote = viper.GetString("app.remote")
	ConfigLocal = viper.GetString("app.local")

	if len(ConfigRemote) == 0 && len(ConfigLocal) == 0 {
		ConfigLocal = UploadRoot
	}
}
