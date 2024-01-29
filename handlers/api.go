package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/lixiang4u/imago/utils"
	"math"
)

func respSuccess(data interface{}, msg ...string) *fiber.Map {
	if len(msg) <= 0 {
		msg = append(msg, "")
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
	return &fiber.Map{
		"code":  1,
		"msg":   msg,
		"debug": debug[0],
	}
}
