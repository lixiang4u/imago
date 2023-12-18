package models

type SiteMap struct {
	OriginSite  string // 原始域名（原图域）
	RewriteSite string // 新域名（新图域）
	LocalPath   string // 本地位置（原图本地位置）
}

type AppConfig struct {
	Id   int64
	Name string
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
