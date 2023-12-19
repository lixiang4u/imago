package utils

import (
	"encoding/json"
	"fmt"
	"github.com/h2non/filetype"
	"github.com/h2non/filetype/types"
	"github.com/lixiang4u/imago/models"
	"log"
	"os"
	"path"
	"path/filepath"
	"time"
)

func GetRemoteLocalFilePath(id, originHost, fileExt string) string {
	var p = path.Join(models.RemoteRoot, originHost, id[:2], fmt.Sprintf("%s.%s", id, fileExt))
	return p
}

func GetOutputFilePath(id, originHost, fileExt string) string {
	var p = path.Join(models.OutputRoot, originHost, id[:2], fmt.Sprintf("%s.%s", id, fileExt))
	return p
}

func GetMetaFilePath(id, originHost string) string {
	var p = path.Join(models.MetaRoot, originHost, id[:2], fmt.Sprintf("%s.json", id))
	return p
}

func LogMeta(id, originHost, source, version string) {
	// ext：表示local/用户给定的域名(不带http)
	var p = GetMetaFilePath(id, originHost)
	_ = os.MkdirAll(path.Dir(p), 0666)
	buf, _ := json.Marshal(models.FileMeta{
		Id:      id,
		Origin:  originHost,
		Url:     source,
		Version: version,
	})
	log.Println("[LogMeta.source]", source)
	_ = os.WriteFile(p, buf, 0644)
}

func GetMeta(id, originHost, source, version string) (meta models.FileMeta, err error) {
	var p = GetMetaFilePath(id, originHost)
	buf, err := os.ReadFile(p)
	log.Println("[GetMeta.p]", p)
	if os.IsNotExist(err) {
		LogMeta(id, originHost, source, version)
		return GetMeta(id, originHost, source, version)
	}
	if err != nil {
		return meta, err
	}
	meta = models.FileMeta{}
	if err = json.Unmarshal(buf, &meta); err != nil {
		LogMeta(id, originHost, source, version)
		return GetMeta(id, originHost, source, version)
	}
	return meta, nil
}
func RemoveMeta(id, originHost string) {
	var p = GetMetaFilePath(id, originHost)
	_ = os.Remove(p)
}

func FileExists(fileName string) bool {
	fi, err := os.Stat(fileName)
	if err != nil {
		return false
	}
	if fi.Size() < 32 {
		return false
	}
	if fi.IsDir() {
		return false
	}
	for i := 0; i < 5; i++ {
		if _, ok := models.LocalCache.Get(fileName); ok {
			time.Sleep(time.Second / 5)
		} else {
			return true
		}
	}
	return false
}

func RemoveCache(p string) {
	files, err := filepath.Glob(p + "*")
	if err != nil {
		log.Println("[RemoveCacheError1]", p, err.Error())
		return
	}
	for _, f := range files {
		if err = os.Remove(f); err != nil {
			log.Println("[RemoveCacheError2]", f, err.Error())
		}
	}
}

func FileSize(fileName string) int64 {
	fi, err := os.Stat(fileName)
	if err != nil {
		return 0
	}
	return fi.Size()
}

func GetFileMIME(fileName string) types.MIME {
	buf, _ := os.ReadFile(fileName)
	kind, _ := filetype.Match(buf)
	return kind.MIME
}
