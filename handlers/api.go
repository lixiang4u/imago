package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/lixiang4u/imago/utils"
)

func respSuccess(data interface{}, msg string) *fiber.Map {
	return &fiber.Map{
		"code": 200,
		"msg":  msg,
		"data": data,
	}
}

func respSuccessList(list interface{}, pager *utils.Pager, msg string) *fiber.Map {
	return &fiber.Map{
		"code":  200,
		"msg":   msg,
		"data":  list,
		"total": pager.Total,
		"page":  pager.Page,
		"limit": pager.Limit,
	}
}

func respError(msg string, data interface{}) *fiber.Map {
	return &fiber.Map{
		"code": 1,
		"msg":  msg,
		"data": data,
	}
}
func respErrorDebug(msg string, debug interface{}) *fiber.Map {
	return &fiber.Map{
		"code":  1,
		"msg":   msg,
		"debug": debug,
	}
}
