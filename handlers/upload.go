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
		return ctx.JSON(respError("访客每日处理图片已达上限"))
	}
	if userId > 0 && count >= int64(models.WebUserShrinkCount) {
		return ctx.JSON(respError("每日处理图片已达上限"))
	}

	fh, err := ctx.FormFile("file")
	if err != nil {
		return ctx.JSON(respError("请求参数错误"))
	}
	// 最大100MB
	if fh.Size > int64(models.MaxUpload) || fh.Size <= 4 {
		return ctx.JSON(respErrorDebug("文件大小异常", fmt.Sprintf("size: %d", fh.Size)))
	}

	f, err := fh.Open()
	if err != nil {
		return ctx.JSON(respErrorDebug("文件类型检测失败", err.Error()))
	}
	defer func() { _ = f.Close() }()

	buf, err := io.ReadAll(f)
	if err != nil {
		return ctx.JSON(respErrorDebug("文件类型检测失败", err.Error()))
	}
	var hash = utils.BytesMd5(append(buf, []byte(models.SECRET_KEY)...))

	var ext = strings.ToLower(strings.Trim(path.Ext(fh.Filename), "."))
	var savedFile = utils.GetUploadFilePath(hash, "", ext)
	if utils.FileSize(savedFile) > 0 {
		// 文件是存在的
		return ctx.JSON(respSuccess(fiber.Map{
			"name": fh.Filename,
			"size": fh.Size,
			"path": savedFile,
			"hash": hash,
		}))
	}

	// https://github.com/h2non/filetype
	var mime = utils.GetBytesMIME(&buf)
	if !slices.Contains(imageMIMETypes, mime.Value) {
		return ctx.JSON(respErrorDebug("文件格式不支持", fmt.Sprintf("mime: %s", mime)))
	}
	if err = os.MkdirAll(path.Dir(savedFile), 0666); err != nil {
		return ctx.JSON(respErrorDebug("文件上传失败", err.Error()))
	}
	if err = ctx.SaveFile(fh, savedFile); err != nil {
		return ctx.JSON(respErrorDebug("文件上传失败", err.Error()))
	}

	return ctx.JSON(respSuccess(fiber.Map{
		"name": fh.Filename,
		"size": fh.Size,
		"path": savedFile,
		"hash": hash,
	}))
}
