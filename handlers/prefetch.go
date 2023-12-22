package handlers

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/lixiang4u/imago/models"
	"github.com/lixiang4u/imago/utils"
	"io/fs"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

func loadRemotePrefetchList() []string {
	file, err := os.Open("prefetch.list")
	defer func() { _ = file.Close() }()

	if err != nil {
		fmt.Println("[no prefetch.list found]")
		return nil
	}

	var prefetchList []string
	var reader = bufio.NewReader(file) // 创建带缓冲的读取器
	for {
		line, err := reader.ReadString('\n') // 逐行读取文件内容
		if err != nil {
			break // 读取完毕或发生错误时退出循环
		}
		line = strings.TrimPrefix(strings.TrimSpace(line), "#")
		if !strings.HasPrefix(line, "http://") || !strings.HasPrefix(line, "https://") {
			continue
		}
		prefetchList = append(prefetchList, line)
	}

	return prefetchList
}

func loadLocalPrefetchList() []string {
	var prefetchList []string
	_ = filepath.Walk(models.LocalConfig.App.Local, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}

		prefetchList = append(prefetchList, strings.TrimSpace(path))

		return nil
	})
	return prefetchList
}

func Prefetch() error {
	log.Println("[prefetch run]")

	var imgConfig = &models.ImageConfig{}
	var appConfig = &models.AppConfig{}

	// 优先加载 prefetch.list 预取资源路径数据，再加载 local 配置目录下资源文件
	var prefetchRemoteList = loadRemotePrefetchList()
	var chRemote = make(chan bool, models.LocalConfig.App.PrefetchThreads)

	log.Println("[prefetch remote]", utils.ToJsonString(fiber.Map{"file": "prefetch.list", "lines": len(prefetchRemoteList), "threads": models.LocalConfig.App.PrefetchThreads}, false))

	for _, prefetchUrl := range prefetchRemoteList {
		chRemote <- true

		go func(prefetchUrl string) {
			if err := parseFileFetchCh(chRemote, prefetchUrl, imgConfig, appConfig); err != nil {
				log.Println("[prefetchUrl error]", prefetchUrl, err.Error())
			}
		}(prefetchUrl)

	}

	// 本地
	var prefetchLocalList = loadLocalPrefetchList()
	var chLocal = make(chan bool, models.LocalConfig.App.PrefetchThreads)

	log.Println("[prefetch local]", utils.ToJsonString(fiber.Map{"file": models.LocalConfig.App.PrefetchThreads, "lines": len(prefetchLocalList), "threads": models.LocalConfig.App.PrefetchThreads}, false))

	for _, prefetchUrl := range prefetchLocalList {
		chLocal <- true

		go func(prefetchUrl string) {
			if err := parseFileFetchCh(chLocal, prefetchUrl, imgConfig, appConfig); err != nil {
				log.Println("[prefetchUrl error]", prefetchUrl, err.Error())
			}
		}(prefetchUrl)

	}

	log.Println("[prefetch done]")
	return nil
}

func parseFileFetchCh(ch chan bool, prefetchUrl string, imgConfig *models.ImageConfig, appConfig *models.AppConfig) error {
	defer func() { _ = <-ch }()

	tmpUrl, err := url.Parse(prefetchUrl)
	if err != nil {
		return err
	}
	if len(tmpUrl.Host) > 0 {
		appConfig.OriginSite = tmpUrl.Host
		appConfig.LocalPath = ""
		appConfig.Refresh = 0
		prefetchUrl = tmpUrl.Path
	} else {
		appConfig.OriginSite = ""
		appConfig.LocalPath = prefetchUrl
		appConfig.Refresh = 0
	}

	localMeta, err := HandleLocalMeta(prefetchUrl, imgConfig, appConfig)
	if err != nil {
		log.Println("[parse meta]", err.Error())
		return err
	}

	log.Println("[debug]", utils.ToJsonString(fiber.Map{
		"prefetchUrl": prefetchUrl,
		"localMeta":   localMeta,
	}, true))

	if utils.FileSize(localMeta.RemoteLocal) > 0 {
		// 文件已经存在
		return errors.New("prefetch file exists")
	}

	if localMeta.Remote {
		// 需要回源，清理老数据
		utils.RemoveCache(localMeta.RemoteLocal)
		utils.RemoveMeta(localMeta.Id, localMeta.Origin)
		utils.LogMeta(localMeta.Id, localMeta.Origin, localMeta.Raw, models.Empty)

		log.Println("[fetch source]", localMeta.Raw, "=>", localMeta.RemoteLocal)

		if err = downloadFile(localMeta.Raw, localMeta.RemoteLocal); err != nil {
			log.Println("[prefetch error]", localMeta.Raw, err.Error())
			return err
		}
	}

	return nil
}
