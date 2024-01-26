package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/lixiang4u/imago/models"
	"github.com/lixiang4u/imago/utils"
	"path"
	"path/filepath"
	"strings"
)

func Download(ctx *fiber.Ctx) error {
	var fileName = ctx.Query("filename", filepath.Base(ctx.Params("*")))
	uploadRoot, err := filepath.Abs(path.Dir(models.UploadRoot))
	if err != nil {
		return ctx.JSON(fiber.Map{
			"error": "上传目录配置异常",
			"debug": err.Error(),
		})
	}
	requestFile, err := filepath.Abs(path.Join(path.Dir(models.UploadRoot), ctx.Params("*")))
	if err != nil {
		return ctx.JSON(fiber.Map{
			"error": "文件不存在",
			"debug": err.Error(),
		})
	}
	if !strings.HasPrefix(requestFile, uploadRoot) {
		return ctx.JSON(fiber.Map{
			"error": "下载文件地址异常",
			"debug": ctx.Params("*"),
		})
	}
	if utils.FileSize(requestFile) <= 0 {
		return ctx.JSON(fiber.Map{
			"error": "文件不存在",
			"debug": ctx.Params("*"),
		})
	}
	return ctx.Download(requestFile, fileName)
}
