package utils

import "net/url"

func ParseUrlHost(tmpUrl string) string {
	u, err := url.Parse(tmpUrl)
	if err != nil {
		return ""
	}
	return u.Host
}
