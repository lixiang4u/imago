package models

type SiteMap struct {
	OriginSite  string // 原始域名（原图域）
	RewriteSite string // 新域名（新图域）
	LocalPath   string // 本地位置（原图本地位置）
}
