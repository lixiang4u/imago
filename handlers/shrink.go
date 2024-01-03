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
	appConfig, err := models.GetHostUserConfig(string(ctx.Request().Host()))
	if err != nil {
		return ctx.JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	fh, err := ctx.FormFile("file")
	if err != nil {
		return ctx.JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// 最大100MB
	if fh.Size > int64(models.MaxUpload) || fh.Size <= 4 {
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

	var fileMIME = utils.GetFileMIME(localMeta.Raw)
	var dstFormat = fileMIME.Subtype
	if len(imgConfig.Format) > 0 {
		dstFormat = imgConfig.Format
	}

	log.Println("[upload file MIME]", utils.ToJsonString(fiber.Map{
		"file": localMeta.Raw,
		"MIME": fileMIME,
	}, false))

	var convertedFile = fmt.Sprintf("%s.%s.%s", localMeta.Raw, localMeta.FeatureId, dstFormat)

	var requestLog = &models.RequestLog{
		UserId:    appConfig.UserId,
		ProxyId:   appConfig.ProxyId,
		MetaId:    localMeta.Id,
		OriginUrl: localMeta.Raw,
		Referer:   ctx.Get("Referer") + ", " + imgConfig.HttpUA,
		Ip:        ctx.IP(),
		IsCache:   0,
		CreatedAt: time.Now(),
	}

	if utils.FileExists(convertedFile) {
		var _size = utils.FileSize(convertedFile)

		go prepareShrinkLog(convertedFile, _size, 1, requestLog, &localMeta, &appConfig)

		return ctx.JSON(fiber.Map{
			"status": "ok",
			"url":    convertedFile,
			"size":   _size,
			"rate":   utils.CompressRate(localMeta.Size, _size),
			"time":   time.Now().Format("2006-01-02 15:04:05"),
		})
	}

	imgConfig.XAutoRotate = true
	_converted, _size, err := ConvertImage(
		localMeta.Raw,
		convertedFile,
		dstFormat,
		&imgConfig,
		&exportConfig,
	)
	if err != nil {
		return ctx.JSON(fiber.Map{
			"error": "failed: " + err.Error(),
		})
	}

	go prepareShrinkLog(convertedFile, _size, 0, requestLog, &localMeta, &appConfig)

	return ctx.JSON(fiber.Map{
		"status": "ok",
		"url":    _converted,
		"size":   _size,
		"rate":   utils.CompressRate(localMeta.Size, _size),
		"time":   time.Now().Format("2006-01-02 15:04:05"),
	})
}

func prepareShrinkLog(convertedFile string, convertedSize int64, isExist int8, requestLog *models.RequestLog, localMeta *models.LocalMeta, appConfig *models.AppConfig) {
	var now = time.Now()

	requestLog.IsExist = isExist
	_ = prepareRequestLog(requestLog, 0, 1)

	_ = prepareRequestStat(&models.RequestStat{
		UserId:       appConfig.UserId,
		ProxyId:      appConfig.ProxyId,
		MetaId:       localMeta.Id,
		OriginUrl:    localMeta.Raw,
		RequestCount: 1,
		ResponseByte: uint64(convertedSize),
		SavedByte:    uint64(localMeta.Size - convertedSize),
		CreatedAt:    now,
	})
	_ = prepareUserFiles(&models.UserFiles{
		UserId:      appConfig.UserId,
		ProxyId:     appConfig.ProxyId,
		MetaId:      localMeta.Id,
		OriginUrl:   localMeta.Raw,
		OriginFile:  localMeta.RemoteLocal,
		ConvertFile: convertedFile,
		OriginSize:  uint64(localMeta.Size),
		ConvertSize: uint64(convertedSize),
		CreatedAt:   now,
	})

}
