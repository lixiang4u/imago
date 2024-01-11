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

type RequestStatRequestChartHandler struct {
	m sync.Map
}

func (x *RequestStatRequestChartHandler) HandleMessage(message *nsq.Message) error {
	if len(message.Body) == 0 {
		return nil
	}
	var tmpRequestStat models.RequestStat
	if err := json.Unmarshal(message.Body, &tmpRequestStat); err != nil {
		return err
	}
	if tmpRequestStat.UserId <= 0 || tmpRequestStat.ProxyId <= 0 {
		return nil
	}
	var tmpRequestStatRequestChart = models.RequestStatRequestChart{
		UserId:       tmpRequestStat.UserId,
		ProxyId:      tmpRequestStat.ProxyId,
		RequestCount: tmpRequestStat.RequestCount,
		ResponseByte: tmpRequestStat.ResponseByte,
		SavedByte:    tmpRequestStat.SavedByte,
		CreatedAt:    tmpRequestStat.CreatedAt.Truncate(time.Minute), // 重新计算时间
	}

	var id uint64
	var mKey = fmt.Sprintf("%d%d%s", tmpRequestStatRequestChart.UserId, tmpRequestStatRequestChart.ProxyId, tmpRequestStatRequestChart.CreatedAt)
	v, ok := x.m.Load(mKey)
	if !ok {
		findRequestStatRequestChart, err := models.GetOrCreateRequestStatRequestChart(tmpRequestStatRequestChart)
		if err != nil {
			log.Println("[GetOrCreateRequestStatRequestChart.Error]", err.Error())
			return err
		}
		id = findRequestStatRequestChart.Id
		x.m.Store(mKey, models.CacheMapValue{Id: id, Timestamp: time.Now().Unix()})
	} else {
		txErr := models.DB().Transaction(func(tx *gorm.DB) error {
			if e := models.DB().Model(&tmpRequestStatRequestChart).Where("id", v.(models.CacheMapValue).Id).Updates(map[string]interface{}{
				"request_count": gorm.Expr("request_count+?", tmpRequestStatRequestChart.RequestCount),
				"response_byte": gorm.Expr("response_byte+?", tmpRequestStatRequestChart.ResponseByte),
				"saved_byte":    gorm.Expr("saved_byte+?", tmpRequestStatRequestChart.SavedByte),
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

func (x *RequestStatRequestChartHandler) FreeCache() {
	var s = uint(0)
	var d = uint(0)
	var ts = time.Now().Unix()
	x.m.Range(func(key, value any) bool {
		if ts-value.(models.CacheMapValue).Timestamp > 3600 || s > 20000 {
			d++
			x.m.Delete(key)
		}
		s++
		return true
	})
	log.Println(fmt.Sprintf("[NsqRequestStat.cache] %d/%d", d, s))
}
