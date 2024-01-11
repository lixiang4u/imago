package handlers

import (
	"fmt"
	"github.com/lixiang4u/imago/handlers/nsq_handler"
	"github.com/lixiang4u/imago/models"
	"github.com/nsqio/go-nsq"
	"log"
	"time"
)

type NsqConsumeHandler struct {
	consumer1 *nsq.Consumer
	consumer2 *nsq.Consumer
	consumer3 *nsq.Consumer
	consumer4 *nsq.Consumer
}

func (x *NsqConsumeHandler) HandleMessage() error {
	var err error
	var h1 = &nsq_handler.RequestHandler{}
	x.consumer1, err = models.NsqConsumer(models.TopicRequest, models.NsqChannel, h1)
	if err != nil {
		return err
	}

	var h2 = &nsq_handler.RequestStatHandler{}
	x.consumer2, err = models.NsqConsumer(models.TopicRequestStat, models.NsqChannel, h2)
	if err != nil {
		return err
	}

	var h3 = &nsq_handler.UserFilesHandler{}
	x.consumer3, err = models.NsqConsumer(models.TopicUserFiles, models.NsqChannel, h3)
	if err != nil {
		return err
	}

	var h4 = &nsq_handler.RequestStatRequestChartHandler{}
	x.consumer4, err = models.NsqConsumer(models.TopicRequestStat, fmt.Sprintf("%s-request-chart", models.NsqChannel), h4)
	if err != nil {
		return err
	}

	go func() {
		var t = time.NewTicker(time.Minute * 5)
		defer t.Stop()
		for {
			select {
			case <-t.C:
				var tp = time.Now().Unix()
				log.Println("[HandleMessage.Ticker]")
				h2.FreeCache()
				h3.FreeCache()
				h4.FreeCache()
				log.Println(fmt.Sprintf("[HandleMessage.Ticker] 耗时 %d 秒", time.Now().Unix()-tp))
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
	if x.consumer4 != nil {
		log.Println("consumer4.Stop()")
		x.consumer4.Stop()
	}
}
