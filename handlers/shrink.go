package handlers

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/lixiang4u/imago/models"
	"github.com/lixiang4u/imago/utils"
	"log"
	"os"
	"path"
	"slices"
	"strings"
	"time"
)

func Shrink(ctx *fiber.Ctx) error {
	var imgConfig = parseConfig(ctx)
	var exportConfig = models.ExportConfig{
		StripMetadata: true,
		Quality:       int(imgConfig.Quality),
		Lossless:      false,
	}

	fh, err := ctx.FormFile("file")
	if err != nil {
		return ctx.JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// 最大100MB
	if fh.Size > int64(models.MAX_UPLOAD) || fh.Size <= 4 {
		return ctx.JSON(fiber.Map{
			"error": fmt.Sprintf("file size irregular(%d)", fh.Size),
		})
	}

	var localMeta = models.LocalMeta{
		Id:        utils.FormattedUUID(16),
		FeatureId: "default",
		Origin:    "",
		Remote:    false,
		Ext:       strings.ToLower(strings.Trim(path.Ext(fh.Filename), ".")),
		Raw:       "",
		Size:      fh.Size,
	}

	localMeta.Raw = utils.GetUploadFilePath(localMeta.Id, localMeta.Origin, localMeta.Ext)
	localMeta.RemoteLocal = localMeta.Raw
	if !utils.IsDefaultObj(imgConfig, []string{"HttpAccept", "HttpUA", "Src"}) {
		localMeta.FeatureId = utils.HashString(fmt.Sprintf("%v", imgConfig))[:6]
	}

	if err = os.MkdirAll(path.Dir(localMeta.Raw), 0666); err != nil {
		return ctx.JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	if !slices.Contains(models.ImageTypes, localMeta.Ext) {
		return ctx.JSON(fiber.Map{
			"error": "file not support",
		})
	}
	if err = ctx.SaveFile(fh, localMeta.Raw); err != nil {
		return ctx.JSON(fiber.Map{
			"error": "upload failed: " + err.Error(),
		})
	}

	log.Println("[upload file MIME]", utils.ToJsonString(fiber.Map{
		"file": localMeta.Raw,
		"MIME": utils.GetFileMIME(localMeta.Raw),
	}, false))

	var convertedFile = fmt.Sprintf(
		"%s.%s.%s",
		utils.GetOutputFilePath(localMeta.Id, localMeta.Origin, localMeta.Ext),
		localMeta.FeatureId,
		localMeta.Ext,
	)

	if utils.FileExists(convertedFile) {
		var _size = utils.FileSize(convertedFile)
		return ctx.JSON(fiber.Map{
			"status": "ok",
			"url":    convertedFile,
			"size":   _size,
			"rate":   utils.CompressRate(localMeta.Size, _size),
			"time":   time.Now().Format("2006-01-02 15:04:05"),
		})
	}

	_converted, _size, err := ConvertImage(
		localMeta.Raw,
		convertedFile,
		models.SUPPORT_TYPE_RAW,
		&imgConfig,
		&exportConfig,
	)
	if err != nil {
		return ctx.JSON(fiber.Map{
			"error": "failed: " + err.Error(),
		})
	}

	return ctx.JSON(fiber.Map{
		"status": "ok",
		"url":    _converted,
		"size":   _size,
		"rate":   utils.CompressRate(localMeta.Size, _size),
		"time":   time.Now().Format("2006-01-02 15:04:05"),
	})
}
