package handlers

import "github.com/gofiber/fiber/v2"

func respSuccess(data interface{}, msg string) *fiber.Map {
	return &fiber.Map{
		"code": 200,
		"msg":  msg,
		"data": data,
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
