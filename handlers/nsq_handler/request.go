package nsq_handler

import (
	"encoding/json"
	"github.com/lixiang4u/imago/models"
	"github.com/nsqio/go-nsq"
)

type RequestHandler struct{}

func (x RequestHandler) HandleMessage(message *nsq.Message) error {
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
