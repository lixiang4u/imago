package models

import (
	"github.com/patrickmn/go-cache"
	"time"
)

var (
	Empty      = ""
	Local      = "local"
	MetaRoot   = "./meta"   //元数据数据存储路径
	RemoteRoot = "./remote" //远程图片原图存储目录
	OutputRoot = "./output" // 压缩等操作后的图片文件
	LocalCache = cache.New(5*time.Minute, 10*time.Minute)
	ImageTypes = []string{"jpg", "png", "jpeg", "bmp", "gif", "svg", "heic"}

	SUPPORT_TYPE_RAW    = "raw"
	SUPPORT_TYPE_WEBP   = "webp"
	SUPPORT_TYPE_AVIF   = "avif"
	SUPPORT_TYPE_JPG    = "jpg"
	SUPPORT_TYPE_NATIVE = "native"
)

type AppConfig struct {
	UserId     string // 用户ID
	OriginSite string //原始域名（原图域）
	LocalPath  string //本地位置（原图本地位置）
	Refresh    int    //是否回源，1.是，0.否
}

// 当前任务相关配置
type ExportConfig struct {
	//StripMetadata：表示是否去除图片的元数据（如Exif信息等）。
	//Quality：表示图片的质量，取值范围为 1 到 100。较高的值表示更好的质量，但文件大小也会增加。
	//Lossless：表示是否使用无损压缩。如果设置为 true，则不会对图片进行任何有损压缩，保持完全的像素精度；如果设置为 false，则会应用有损压缩以减小文件大小。
	//NearLossless：表示是否使用近似无损压缩。如果设置为 true，则会应用一定程度的有损压缩，但保持较高的视觉质量；如果设置为 false，则不会应用近似无损压缩。
	//ReductionEffort：表示压缩的努力程度。取值范围为 0 到 6，0 表示最快压缩速度但压缩率较低，6 表示最慢压缩速度但压缩率较高。
	StripMetadata   bool
	Quality         int
	Lossless        bool
	NearLossless    bool
	ReductionEffort int
}

type ImageConfig struct {
	HttpAccept string
	HttpUA     string
	Src        string
	Refresh    int
	With       float64
	Height     float64
	Flip       string
	Quality    float64
	Blur       float64
	Sharpen    float64
	Rotate     float64
	Brightness float64
	Saturation float64
	Hue        float64
	Contrast   float64
	Watermark  struct {
		Text    string
		Font    string
		Color   string
		With    float64
		Height  float64
		OffsetX float64
		OffsetY float64
		Opacity float64
	}
	Filter string
}

type FileMeta struct {
	Id      string // 文件ID
	Origin  string // local/用户配置的原创图片域名
	Url     string // 本地路径/远程URL
	Version string // 文件版本？
}

type LocalMeta struct {
	Id          string
	Origin      string
	Remote      bool
	Ext         string
	RemoteLocal string
	Raw         string
	Size        int64
}
