package utils

import (
	"github.com/gofiber/fiber/v2"
	"log"
	"net/http"
)

func GetResourceVersion(requestUrl string, keys []string) string {
	// 可以自定义配置，例如：Content-Md5
	if len(keys) == 0 {
		keys = []string{"Etag", "Content-Length", "Content-Type"}
	}
	resp, err := http.Head(requestUrl)
	if err != nil {
		log.Println("[http.Head]", err.Error())
		return ""
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	log.Println("[fetch resource]", ToJsonString(fiber.Map{"content_type": resp.Header.Get("Content-Type"), "requestUrl": requestUrl}, false))

	var s = ""
	for _, key := range keys {
		s += resp.Header.Get(key)
	}
	return s
}
