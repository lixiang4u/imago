package nsq_handler

import (
	"encoding/json"
	"github.com/lixiang4u/imago/models"
	"github.com/nsqio/go-nsq"
	"log"
)

type AdminCommandHandler struct {
	//m sync.Map
}

func (x *AdminCommandHandler) HandleMessage(message *nsq.Message) error {
	if len(message.Body) == 0 {
		return nil
	}
	var cmd models.AdminCommand
	if err := json.Unmarshal(message.Body, &cmd); err != nil {
		return err
	}
	switch cmd.Command {
	case 0x0010: // 后台更新用户代理触发
		x.cmdUpdateUserProxy(cmd.Body)
	case 0x0020:
	case 0x0030:
	case 0x0040:
	case 0x0050:
	case 0x0060:
	case 0x0070:
	case 0x0080:
	case 0x0090:
	default:
		log.Println("[command not match]", cmd.Command)
	}

	return nil
}

func (x *AdminCommandHandler) FreeCache() {}

func (x *AdminCommandHandler) cmdUpdateUserProxy(body interface{}) {
	up, ok := body.(models.UserProxy)
	if !ok {
		log.Println("[cmdUpdateUserProxy] params type error")
		return
	}
	models.LocalCache.Delete(models.GetHostCacheKey(up.Host))
	models.LocalCache.Delete(models.GetHostCacheKey(up.Host, up.UserId))

	log.Println("[cmdUpdateUserProxy] ok: ", up.UserId, up.Host)
}
