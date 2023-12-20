package utils

import (
	"github.com/gofiber/fiber/v2"
	"github.com/lixiang4u/imago/models"
	"log"
	"net/http"
)

func GetResourceVersion(requestUrl string, keys []string) string {
	// 可以自定义配置，例如：Content-Md5
	if len(keys) == 0 {
		keys = []string{"Etag", "Content-Length", "Content-Type"}
	}
	req, err := http.NewRequest("HEAD", requestUrl, nil)
	if err != nil {
		log.Println("[http.head]", err.Error())
		return ""
	}
	req.Header.Set("User-Agent", models.UserAgent)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("[http.head]", err.Error())
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
