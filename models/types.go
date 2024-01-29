package models

type AppConfig struct {
	UserId     uint64 // 用户ID
	ProxyId    uint64 // 代理ID
	OriginSite string //原始域名（原图域）
	LocalPath  string //本地位置（原图本地位置）
	ProxyHost  string //代理域名（用户请求代理域名，最终源文件在OriginSite）
	Refresh    int    //是否回源，1.是，0.否
	Debug      bool   //是否调试模式
	UserAgent  string //溯源用的User-Agent
	Cors       string //跨域设置
}

type ExportConfig struct {
	//StripMetadata：表示是否去除图片的元数据（如Exif信息等）。
	//Quality：表示图片的质量，取值范围为 1 到 100。较高的值表示更好的质量，但文件大小也会增加。
	//Lossless：表示是否使用无损压缩。如果设置为 true，则不会对图片进行任何有损压缩，保持完全的像素精度；如果设置为 false，则会应用有损压缩以减小文件大小。
	//NearLossless：表示是否使用近似无损压缩。如果设置为 true，则会应用一定程度的有损压缩，但保持较高的视觉质量；如果设置为 false，则不会应用近似无损压缩。
	//ReductionEffort：表示压缩的努力程度。取值范围为 0 到 6，0 表示最快压缩速度但压缩率较低，6 表示最慢压缩速度但压缩率较高。
	//OptimizeCoding：是否使用优化编码，如果为 true，则使用优化编码；如果为 false，则使用普通编码。
	//Compression：PNG 图像的压缩级别，取值范围一般为 0 到 9，数值越大表示压缩率越高但速度越慢。
	StripMetadata   bool
	Quality         int
	Lossless        bool
	NearLossless    bool
	ReductionEffort int
	OptimizeCoding  bool
	Compression     int
}

type ImageConfig struct {
	HttpAccept   string   `json:"http_accept" form:"http_accept"`     // 特殊字段，从http头获取
	HttpUA       string   `json:"http_ua" form:"http_ua"`             // 特殊字段，从http头获取
	XAutoRotate  bool     `json:"x_auto_rotate" form:"x_auto_rotate"` // 特殊字段，是否AutoRotate
	Src          string   `json:"src" form:"src"`                     // 特殊字段，表示目标图片url
	Refresh      int      `json:"refresh" form:"refresh"`             // 特殊字段，1.表示强制回源，0.默认不强制回源
	Width        float64  `json:"width" form:"width"`                 // 图片resize的宽度
	Height       float64  `json:"height" form:"height"`               // 图片resize的高度
	Flip         string   `json:"flip" form:"flip"`                   // Available options are: v(vertical), h(horizontal), b(Both vertical and horizontal)
	Quality      float64  `json:"quality" form:"quality"`             // Override quality set in dashbaord, available quality range from 10 ~ 100(100 means lossless convert)
	Blur         float64  `json:"blur" form:"blur"`                   // Available blur range from 10 ~ 100
	Sharpen      float64  `json:"sharpen" form:"sharpen"`             // Sharpen the image, available sharpen range from 1 ~ 10
	Rotate       float64  `json:"rotate" form:"rotate"`               // Available rotate angle range from 0 ~ 360, however if angle is not 90, 180, 270, 360, it will be filled with white background
	Brightness   float64  `json:"brightness" form:"brightness"`       // Adjust brightness of the image, available range from 0 ~ 10, 1 means no change
	Saturation   float64  `json:"saturation" form:"saturation"`       // Adjust saturation of the image, available range from 0 ~ 10, 1 means no change
	Hue          float64  `json:"hue" form:"hue"`                     // Adjust hue of the image, available range from 0 ~ 360, hue will be 0 for no change, 90 for a complementary hue shift, 180 for a contrasting shift, 360 for no change again.
	Contrast     float64  `json:"contrast" form:"contrast"`           // Adjust contrast of the image, available range from 0 ~ 10, 1 means no change
	VisualEffect []string `json:"visual_effect" form:"visual_effect"` // 图片添加filter和水印相关，需要编码/解码
	Format       string   `json:"format" form:"format"`               // 目标图格式，SUPPORT_TYPE_*
	Compression  int      `json:"compression" form:"compression"`     // PNG 文件压缩级别，取值范围为 0~9，0 表示无压缩，9 表示最高压缩率。
}

type Watermark struct {
	Text    string  `json:"text" form:"text"`
	Font    string  `json:"font" form:"font"`
	Color   string  `json:"color" form:"color"`
	Width   float64 `json:"width" form:"width"`
	Height  float64 `json:"height" form:"height"`
	OffsetX float64 `json:"offset_x" form:"offset_x"`
	OffsetY float64 `json:"offset_y" form:"offset_y"`
	Opacity float64 `json:"opacity" form:"opacity"`
	Align   string  `json:"align" form:"align"`
}

type FileMeta struct {
	Id      string // 文件ID
	Origin  string // local/用户配置的原创图片域名
	Url     string // 本地路径/远程URL
	Version string // 文件版本？
}

type LocalMeta struct {
	Id          string
	FeatureId   string // 功能参数后的编码id
	Origin      string // 源主机（不带协议）
	Remote      bool   // 源文件是否远程路径
	Ext         string // 源文件后缀
	RemoteLocal string // 远程源文件在本地地址
	Raw         string // 源文件的位置（本地或者远程URL）
	RawVersion  string
	RequestUri  string // 请求的路径（带参数）
	Size        int64  // 原始文件大小
	FetchSource bool   // 是否溯源
}

type LocalAppConfig struct {
	App struct {
		Remote          string   `json:"remote"`
		Local           string   `json:"local"`
		Prefetch        bool     `json:"prefetch"`
		PrefetchThreads int      `json:"prefetch_threads"`
		Transform       []string `json:"transform"`
	} `json:"app"`
	MySQL struct {
		Dsn string `json:"dsn"`
	} `json:"mysql"`
	// 内存数据
	L struct {
		Supported map[string]bool
	} `json:"l"`
}

type CacheMapValue struct {
	Id        uint64
	Timestamp int64
}

type SimpleFile struct {
	Path string `json:"path" form:"path"`
	Name string `json:"name" form:"name"`
}

type AdminCommand struct {
	Command int         `json:"command" form:"command"`
	Body    interface{} `json:"body" form:"body"`
}
