package handlers

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/lixiang4u/imago/models"
	"github.com/lixiang4u/imago/utils"
	"io"
	"os"
	"path"
	"slices"
	"strings"
)

var imageMIMETypes = []string{"image/jpeg", "image/png", "image/gif", "image/webp", "image/bmp", "image/avif", "image/heif"}

func Upload(ctx *fiber.Ctx) error {
	var userId = utils.TryParseUserId(ctx)
	var count = models.IncrementWebShrinkCount(utils.GetClientIp(ctx))
	if userId <= 0 && count >= int64(models.GuestUserShrinkCount) {
		return ctx.JSON(fiber.Map{
			"error": "访客每日处理图片已达上限",
		})
	}
	if userId > 0 && count >= int64(models.WebUserShrinkCount) {
		return ctx.JSON(fiber.Map{
			"error": "每日处理图片已达上限",
		})
	}

	fh, err := ctx.FormFile("file")
	if err != nil {
		return ctx.JSON(fiber.Map{
			"error": "请求参数错误",
			"debug": err.Error(),
		})
	}
	// 最大100MB
	if fh.Size > int64(models.MaxUpload) || fh.Size <= 4 {
		return ctx.JSON(fiber.Map{
			"error": "文件大小异常",
			"debug": fmt.Sprintf("size: %d", fh.Size),
		})
	}

	var ext = strings.ToLower(strings.Trim(path.Ext(fh.Filename), "."))
	var savedFile = utils.GetUploadFilePath(utils.FormattedUUID(32), "", ext)

	if err = os.MkdirAll(path.Dir(savedFile), 0666); err != nil {
		return ctx.JSON(fiber.Map{
			"error": "文件上传失败",
			"debug": err.Error(),
		})
	}
	f, err := fh.Open()
	if err != nil {
		return ctx.JSON(fiber.Map{
			"error": "文件类型检测失败",
			"debug": err.Error(),
		})
	}
	defer func() { _ = f.Close() }()

	buf, err := io.ReadAll(f)
	if err != nil {
		return ctx.JSON(fiber.Map{
			"error": "文件类型检测失败",
			"debug": err.Error(),
		})
	}
	var hash = utils.BytesMd5(buf)

	// https://github.com/h2non/filetype
	var mime = utils.GetReaderMIME(f)
	if !slices.Contains(imageMIMETypes, mime.Value) {
		return ctx.JSON(fiber.Map{
			"error": "文件格式不支持",
			"mime":  mime,
		})
	}
	if err = ctx.SaveFile(fh, savedFile); err != nil {
		return ctx.JSON(fiber.Map{
			"error": "文件上传失败",
			"debug": err.Error(),
		})
	}

	return ctx.JSON(fiber.Map{
		"status": "ok",
		"name":   fh.Filename,
		"size":   fh.Size,
		"path":   savedFile,
		"hash":   hash,
	})
}
