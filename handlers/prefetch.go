package handlers

import (
	"bufio"
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
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "http://") || !strings.HasPrefix(line, "https://") {
			continue
		}
		prefetchList = append(prefetchList)
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

		prefetchList = append(prefetchList, path)

		return nil
	})
	return prefetchList
}

func Prefetch() error {
	log.Println("[prefetch run]")

	// 优先加载 prefetch.list 预取资源路径数据，再加载 local 配置目录下资源文件
	var prefetchRemoteList = loadRemotePrefetchList()

	log.Println("[prefetch remote]", utils.ToJsonString(fiber.Map{"file": "prefetch.list", "lines": len(prefetchRemoteList), "threads": models.LocalConfig.App.PrefetchThreads}, false))

	var chRemote = make(chan bool, models.LocalConfig.App.PrefetchThreads)
	for _, prefetchUrl := range prefetchRemoteList {
		chRemote <- true
		go func(ch chan bool, prefetchUrl string) {
			//任务，结束后  <-ch 释放chan空间
			defer func() { _ = <-ch }()

			_, fileExt, ok := CheckFileAllowed(prefetchUrl)
			if !ok {
				return
			}

			var rawFile = models.Empty
			var rawFileClean = models.Empty
			tmpUrl, err := url.Parse(prefetchUrl)
			if err != nil {
				return
			}
			rawFile = fmt.Sprintf("%s/%s", strings.TrimRight(tmpUrl.Host, "/"), strings.TrimLeft(tmpUrl.Path, "/"))
			tmpRawUrl, err := url.Parse(rawFile)
			if err != nil {
				rawFileClean = tmpUrl.Path
			} else {
				rawFileClean = tmpRawUrl.Path
			}
			var id = utils.HashString(fmt.Sprintf("%s,%s", tmpUrl.Host, rawFileClean))
			var featureId = "default"

			var localMeta = models.LocalMeta{
				Id:          id,
				FeatureId:   featureId,
				Remote:      true,
				Origin:      tmpUrl.Host,
				Ext:         fileExt,
				RemoteLocal: rawFile,
				Raw:         rawFile,
				Size:        0,
			}

			localMeta.RemoteLocal = utils.GetRemoteLocalFilePath(id, localMeta.Origin, fileExt)

			if utils.FileSize(localMeta.RemoteLocal) > 0 {
				// 文件已经存在
				return
			}

			// 需要回源，清理老数据
			utils.RemoveCache(localMeta.RemoteLocal)
			utils.RemoveMeta(id, localMeta.Origin)
			utils.LogMeta(id, localMeta.Origin, rawFile, models.Empty)

			log.Println("[fetch source]", rawFile, "=>", localMeta.RemoteLocal)

			if err = downloadFile(rawFile, localMeta.RemoteLocal); err != nil {
				log.Println("[prefetch error]", rawFile, err.Error())
				return
			}

		}(chRemote, prefetchUrl)
	}

	// 本地
	var prefetchLocalList = loadLocalPrefetchList()

	log.Println("[prefetch local]", utils.ToJsonString(fiber.Map{"file": models.LocalConfig.App.PrefetchThreads, "lines": len(prefetchLocalList), "threads": models.LocalConfig.App.PrefetchThreads}, false))

	var chLocal = make(chan bool, models.LocalConfig.App.PrefetchThreads)
	for _, prefetchUrl := range prefetchLocalList {
		chLocal <- true
		go func(ch chan bool, prefetchUrl string) {
			//任务，结束后  <-ch 释放chan空间
			defer func() { _ = <-ch }()

			_, fileExt, ok := CheckFileAllowed(prefetchUrl)
			if !ok {
				return
			}

			var rawFile = models.Empty
			var rawFileClean = models.Empty
			tmpUrl, err := url.Parse(prefetchUrl)
			if err != nil {
				return
			}
			rawFile = fmt.Sprintf("%s/%s", strings.TrimRight(tmpUrl.Host, "/"), strings.TrimLeft(tmpUrl.Path, "/"))
			tmpRawUrl, err := url.Parse(rawFile)
			if err != nil {
				rawFileClean = tmpUrl.Path
			} else {
				rawFileClean = tmpRawUrl.Path
			}
			var id = utils.HashString(fmt.Sprintf("%s,%s", tmpUrl.Host, rawFileClean))
			var featureId = "default"

			var localMeta = models.LocalMeta{
				Id:          id,
				FeatureId:   featureId,
				Remote:      true,
				Origin:      tmpUrl.Host,
				Ext:         fileExt,
				RemoteLocal: rawFile,
				Raw:         rawFile,
				Size:        0,
			}

			localMeta.RemoteLocal = utils.GetRemoteLocalFilePath(id, localMeta.Origin, fileExt)

			if utils.FileSize(localMeta.RemoteLocal) > 0 {
				// 文件已经存在
				return
			}

			// 需要回源，清理老数据
			utils.RemoveCache(localMeta.RemoteLocal)
			utils.RemoveMeta(id, localMeta.Origin)
			utils.LogMeta(id, localMeta.Origin, rawFile, models.Empty)

			log.Println("[fetch source]", rawFile, "=>", localMeta.RemoteLocal)

			if err = downloadFile(rawFile, localMeta.RemoteLocal); err != nil {
				log.Println("[prefetch error]", rawFile, err.Error())
				return
			}

		}(chLocal, prefetchUrl)
	}

	log.Println("[prefetch done]")
	return nil
}
