package nsq_handler

import (
	"encoding/json"
	"fmt"
	"github.com/lixiang4u/imago/models"
	"github.com/nsqio/go-nsq"
	"log"
	"sync"
	"time"
)

type UserFilesHandler struct {
	m sync.Map
}

func (x *UserFilesHandler) HandleMessage(message *nsq.Message) error {
	if len(message.Body) == 0 {
		return nil
	}
	var tmpUserFiles models.UserFiles
	if err := json.Unmarshal(message.Body, &tmpUserFiles); err != nil {
		return err
	}
	if tmpUserFiles.UserId <= 0 || tmpUserFiles.ProxyId <= 0 || len(tmpUserFiles.MetaId) == 0 {
		return nil
	}
	var mKey = fmt.Sprintf("%d%d%s", tmpUserFiles.UserId, tmpUserFiles.ProxyId, tmpUserFiles.MetaId)
	if _, ok := x.m.Load(mKey); ok {
		return nil
	}
	findUserFiles, err := models.GetOrCreateUserFiles(tmpUserFiles)
	if err != nil {
		log.Println("[GetOrCreateUserFiles.Error]", err.Error())
		return err
	}
	x.m.Store(mKey, models.CacheMapValue{Id: findUserFiles.Id, Timestamp: time.Now().Unix()})

	return nil
}

func (x *UserFilesHandler) FreeCache() {
	var s = uint(0)
	var d = uint(0)
	var a = uint(0)

	var ts = time.Now().Unix()
	x.m.Range(func(key, value any) bool {
		if ts-value.(models.CacheMapValue).Timestamp > 86400 || a > 20000 {
			d++
			x.m.Delete(key)
		} else {
			a++
		}
		s++
		return true
	})
	log.Println(fmt.Sprintf("[NsqUserFiles.cache] %d/%d", d, s))
}
