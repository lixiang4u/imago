package handlers

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

type CacheMapValue struct {
	Id        uint64
	Timestamp int64
}

type NsqConsumeHandler struct {
	consumer1 *nsq.Consumer
	consumer2 *nsq.Consumer
	consumer3 *nsq.Consumer
	consumer4 *nsq.Consumer
}
type NsqRequestHandler struct{}
type NsqRequestStatHandler struct {
	m sync.Map
}
type NsqUserFilesHandler struct {
	m sync.Map
}
type NsqRequestStatRequestChartHandler struct {
	m sync.Map
}

func (x *NsqConsumeHandler) HandleMessage() error {
	var err error
	var h1 = &NsqRequestHandler{}
	x.consumer1, err = models.NsqConsumer(models.TopicRequest, models.NsqChannel, h1)
	if err != nil {
		return err
	}

	var h2 = &NsqRequestStatHandler{}
	x.consumer2, err = models.NsqConsumer(models.TopicRequestStat, models.NsqChannel, h2)
	if err != nil {
		return err
	}

	var h3 = &NsqUserFilesHandler{}
	x.consumer3, err = models.NsqConsumer(models.TopicUserFiles, models.NsqChannel, h3)
	if err != nil {
		return err
	}

	var h4 = &NsqRequestStatRequestChartHandler{}
	x.consumer3, err = models.NsqConsumer(models.TopicRequestStat, fmt.Sprintf("%s-request-chart", models.NsqChannel), h4)
	if err != nil {
		return err
	}

	go func() {
		var t = time.NewTicker(time.Minute * 2)
		defer t.Stop()
		for {
			select {
			case <-t.C:
				log.Println("[HandleMessage.Ticker]")
				h2.freeCache()
				h3.freeCache()
			}
		}
	}()

	return nil
}

func (x *NsqConsumeHandler) NsqStop() {
	if x.consumer1 != nil {
		log.Println("consumer1.Stop()")
		x.consumer1.Stop()
	}
	if x.consumer2 != nil {
		log.Println("consumer2.Stop()")
		x.consumer2.Stop()
	}
	if x.consumer3 != nil {
		log.Println("consumer3.Stop()")
		x.consumer3.Stop()
	}
}

func (x NsqRequestHandler) HandleMessage(message *nsq.Message) error {
	if len(message.Body) == 0 {
		return nil
	}
	var tmpRequestLog models.RequestLog
	if err := json.Unmarshal(message.Body, &tmpRequestLog); err != nil {
		return err
	}
	if tmpRequestLog.UserId <= 0 || tmpRequestLog.ProxyId <= 0 || len(tmpRequestLog.MetaId) == 0 {
		return nil
	}
	if err := models.DB().Model(&tmpRequestLog).Create(&tmpRequestLog).Error; err != nil {
		return err
	}

	return nil
}

func (x *NsqRequestStatHandler) HandleMessage(message *nsq.Message) error {
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
		x.m.Store(mKey, CacheMapValue{Id: id, Timestamp: time.Now().Unix()})
	} else {
		txErr := models.DB().Transaction(func(tx *gorm.DB) error {
			if e := models.DB().Model(&tmpRequestStat).Where("id", v.(CacheMapValue).Id).Updates(map[string]interface{}{
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

func (x *NsqRequestStatHandler) freeCache() {
	var s = uint(0)
	var d = uint(0)
	var ts = time.Now().Unix()
	x.m.Range(func(key, value any) bool {
		if ts-value.(CacheMapValue).Timestamp > 86400 {
			d++
			x.m.Delete(key)
		}
		s++
		return true
	})
	log.Println(fmt.Sprintf("[NsqRequestStat.cache] %d/%d", d, s))
}

func (x *NsqUserFilesHandler) HandleMessage(message *nsq.Message) error {
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
	x.m.Store(mKey, CacheMapValue{Id: findUserFiles.Id, Timestamp: time.Now().Unix()})

	return nil
}

func (x *NsqUserFilesHandler) freeCache() {
	var s = uint(0)
	var d = uint(0)
	var a = uint(0)
	var ts = time.Now().Unix()
	x.m.Range(func(key, value any) bool {
		if ts-value.(CacheMapValue).Timestamp > 86400 || a > 20000 {
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

func (x *NsqRequestStatRequestChartHandler) HandleMessage(message *nsq.Message) error {
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
		x.m.Store(mKey, CacheMapValue{Id: id, Timestamp: time.Now().Unix()})
	} else {
		txErr := models.DB().Transaction(func(tx *gorm.DB) error {
			if e := models.DB().Model(&tmpRequestStatRequestChart).Where("id", v.(CacheMapValue).Id).Updates(map[string]interface{}{
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

func (x *NsqRequestStatRequestChartHandler) freeCache() {
	var s = uint(0)
	var d = uint(0)
	var ts = time.Now().Unix()
	x.m.Range(func(key, value any) bool {
		if ts-value.(CacheMapValue).Timestamp > 3600 || s > 20000 {
			d++
			x.m.Delete(key)
		}
		s++
		return true
	})
	log.Println(fmt.Sprintf("[NsqRequestStat.cache] %d/%d", d, s))
}
