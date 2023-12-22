package models

import (
	"encoding/json"
	"github.com/patrickmn/go-cache"
	"github.com/spf13/viper"
	"log"
	"os"
	"time"
)

var (
	Empty        = ""
	Local        = "local"
	MetaRoot     = "./meta"              //元数据数据存储路径
	RemoteRoot   = "./remote"            //远程图片原图存储目录
	OutputRoot   = "./output"            // 压缩等操作后的图片文件
	UploadRoot   = "./upload"            // web端上传目录
	MaxUpload    = 1 * 1024 * 1024 * 100 // 最大上传文件大小
	LocalCache   = cache.New(5*time.Minute, 10*time.Minute)
	ImageTypes   = []string{"jpg", "png", "jpeg", "bmp", "gif", "svg", "heic", "webp"}
	UserAgent    = "Imago Service/1.0 (89f882e4f6ce47b8)"
	MaxWebpPixel = 16383 // WebP is bitstream-compatible with VP8 and uses 14 bits for width and height. The maximum pixel dimensions of a WebP image is 16383 x 16383.

	SUPPORT_TYPE_RAW    = "raw"
	SUPPORT_TYPE_WEBP   = "webp"
	SUPPORT_TYPE_AVIF   = "avif"
	SUPPORT_TYPE_JPG    = "jpg"
	SUPPORT_TYPE_BMP    = "bmp"
	SUPPORT_TYPE_GIF    = "gif"
	SUPPORT_TYPE_HEIF   = "heif"
	SUPPORT_TYPE_JPEG   = "jpeg"
	SUPPORT_TYPE_PNG    = "png"
	SUPPORT_TYPE_NATIVE = "native"
)
var (
	ConfigRemote string
	ConfigLocal  string
)

func init() {
	LocalConfig.App.Local = UploadRoot

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

	if err = viper.Unmarshal(&LocalConfig); err != nil {
		log.Println("unmarshal config.toml, ", err.Error())
		return
	}

	buf, _ := json.Marshal(LocalConfig)
	log.Println("[LocalConfig]", string(buf))

	if len(LocalConfig.App.Remote) == 0 && len(LocalConfig.App.Local) == 0 {
		LocalConfig.App.Local = UploadRoot
	}
}
