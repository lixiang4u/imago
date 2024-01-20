package utils

import (
	"net/url"
)

func ParseUrlHost(tmpUrl string) string {
	u, err := url.Parse(tmpUrl)
	if err != nil {
		return ""
	}
	return u.Host
}

func ParseUrlDefaultHost(tmpUrl, defaultHost string) string {
	var p = ParseUrlHost(tmpUrl)
	if len(p) == 0 {
		return defaultHost
	}
	return p
}
