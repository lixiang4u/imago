package models

import (
	"errors"
	"github.com/nsqio/go-nsq"
	"log"
)

var (
	TopicRequest      = "topic-request"
	TopicRequestStat  = "topic-request-stat"
	TopicUserFiles    = "topic-user-files"
	TopicAdminCommand = "topic-admin-command" // 管理端修改用户相关信息等，通知服务更新数据
	NsqChannel        = "channel"
	nsqLookupdAddr    = "localhost:4161"
	nsqdAddr          = "127.0.0.1:4150"
	nsqConfig         = nsq.NewConfig()
	nsqProducer       *nsq.Producer
)

func NsqConsumer(topic, channel string, handler nsq.Handler) (consumer *nsq.Consumer, err error) {
	consumer, err = nsq.NewConsumer(topic, channel, nsqConfig)
	if err != nil {
		return
	}

	consumer.AddHandler(handler)

	err = consumer.ConnectToNSQLookupd(nsqLookupdAddr)
	if err != nil {
		return
	}

	return consumer, nil
}

func GetNsqProducer() *nsq.Producer {
	if nsqProducer != nil {
		return nsqProducer
	}
	var err error
	nsqProducer, err = nsq.NewProducer(nsqdAddr, nsqConfig)
	if err != nil {
		log.Println("[nsq.NewProducer]", err.Error())
		nsqProducer = nil
	}
	return nsqProducer
}

func NsqProducer(topic string, message []byte) error {
	var producer = GetNsqProducer()
	if producer == nil {
		return errors.New("nsq producer 创建失败")
	}
	return producer.Publish(topic, message)
}

func NsqProducerClose() {
	if nsqProducer != nil {
		nsqProducer.Stop()
	}
}
