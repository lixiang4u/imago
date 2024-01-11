package nsq_handler

import (
	"encoding/json"
	"fmt"
	"github.com/lixiang4u/imago/models"
	"github.com/nsqio/go-nsq"
	"gorm.io/gorm"
	"log"
	"sync"
	"time"
)

type RequestStatHandler struct {
	m sync.Map
}

func (x *RequestStatHandler) HandleMessage(message *nsq.Message) error {
	if len(message.Body) == 0 {
		return nil
	}
	var tmpRequestStat models.RequestStat
	if err := json.Unmarshal(message.Body, &tmpRequestStat); err != nil {
		return err
	}
	if tmpRequestStat.UserId <= 0 || tmpRequestStat.ProxyId <= 0 || len(tmpRequestStat.MetaId) == 0 {
		return nil
	}
	var id uint64
	var mKey = fmt.Sprintf("%d%d%s", tmpRequestStat.UserId, tmpRequestStat.ProxyId, tmpRequestStat.MetaId)
	v, ok := x.m.Load(mKey)
	if !ok {
		findRequestStat, err := models.GetOrCreateRequestStat(tmpRequestStat)
		if err != nil {
			log.Println("[GetOrCreateRequestStat.Error]", err.Error())
			return err
		}
		id = findRequestStat.Id
		x.m.Store(mKey, models.CacheMapValue{Id: id, Timestamp: time.Now().Unix()})
	} else {
		txErr := models.DB().Transaction(func(tx *gorm.DB) error {
			if e := models.DB().Model(&tmpRequestStat).Where("id", v.(models.CacheMapValue).Id).Updates(map[string]interface{}{
				"request_count": gorm.Expr("request_count+?", tmpRequestStat.RequestCount),
				"response_byte": gorm.Expr("response_byte+?", tmpRequestStat.ResponseByte),
				"saved_byte":    gorm.Expr("saved_byte+?", tmpRequestStat.SavedByte),
			}).Error; e != nil {
				return e
			}
			return nil
		})
		if txErr != nil {
			log.Println("[txErr.Error]", txErr.Error())
			return txErr
		}
	}

	return nil
}

func (x *RequestStatHandler) FreeCache() {
	var s = uint(0)
	var d = uint(0)
	var ts = time.Now().Unix()
	x.m.Range(func(key, value any) bool {
		if ts-value.(models.CacheMapValue).Timestamp > 86400 {
			d++
			x.m.Delete(key)
		}
		s++
		return true
	})
	log.Println(fmt.Sprintf("[NsqRequestStat.cache] %d/%d", d, s))
}
