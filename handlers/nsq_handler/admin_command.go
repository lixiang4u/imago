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
	case models.NsqCmd0x0010:
		x.cmdUpdateUserProxy(cmd.Body)
	case models.NsqCmd0x0020:
	case models.NsqCmd0x0030:
	case models.NsqCmd0x0040:
	case models.NsqCmd0x0050:
	case models.NsqCmd0x0060:
	case models.NsqCmd0x0070:
	case models.NsqCmd0x0080:
	case models.NsqCmd0x0090:
	default:
		log.Println("[command not match]", cmd.Command)
	}

	return nil
}

func (x *AdminCommandHandler) FreeCache() {}

func (x *AdminCommandHandler) cmdUpdateUserProxy(body interface{}) {
	var up models.UserProxy
	err := json.Unmarshal([]byte(body.(string)), &up)
	if err != nil {
		log.Println("[cmdUpdateUserProxy] parseError", body)
		return
	}
	models.LocalCache.Delete(models.GetHostCacheKey(up.Host))
	models.LocalCache.Delete(models.GetHostCacheKey(up.Host, up.UserId))

	log.Println("[cmdUpdateUserProxy] ok: ", up.UserId, up.Host)
}
