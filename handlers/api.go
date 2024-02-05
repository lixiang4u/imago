package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/lixiang4u/imago/utils"
	"log"
	"math"
)

const debugResp = false

func respSuccess(data interface{}, msg ...string) *fiber.Map {
	if len(msg) <= 0 {
		msg = append(msg, "")
	}
	if debugResp {
		log.Println("[respSuccess]", utils.ToJsonString(fiber.Map{
			"code": 200,
			"msg":  msg[0],
			"data": data,
		}, true))
	}
	return &fiber.Map{
		"code": 200,
		"msg":  msg[0],
		"data": data,
	}
}

func respSuccessList(list interface{}, pager *utils.Pager, msg ...string) *fiber.Map {
	if len(msg) <= 0 {
		msg = append(msg, "")
	}
	if debugResp {
		log.Println("[respSuccessList]", utils.ToJsonString(fiber.Map{
			"code":  200,
			"msg":   msg[0],
			"data":  list,
			"total": pager.Total,
			"page":  pager.Page,
			"pages": math.Ceil(float64(pager.Total) / float64(pager.Limit)),
			"limit": pager.Limit,
		}, true))
	}
	return &fiber.Map{
		"code":  200,
		"msg":   msg[0],
		"data":  list,
		"total": pager.Total,
		"page":  pager.Page,
		"pages": math.Ceil(float64(pager.Total) / float64(pager.Limit)),
		"limit": pager.Limit,
	}
}

func respError(msg string, data ...interface{}) *fiber.Map {
	if len(data) <= 0 {
		data = append(data, nil)
	}
	if debugResp {
		log.Println("[respError]", utils.ToJsonString(fiber.Map{
			"code": 1,
			"msg":  msg,
			"data": data[0],
		}, true))
	}
	return &fiber.Map{
		"code": 1,
		"msg":  msg,
		"data": data[0],
	}
}
func respErrorDebug(msg string, debug ...interface{}) *fiber.Map {
	if len(debug) <= 0 {
		debug = append(debug, nil)
	}
	if debugResp {
		log.Println("[respErrorDebug]", utils.ToJsonString(fiber.Map{
			"code":  1,
			"msg":   msg,
			"debug": debug[0],
		}, true))
	}
	return &fiber.Map{
		"code":  1,
		"msg":   msg,
		"debug": debug[0],
	}
}
