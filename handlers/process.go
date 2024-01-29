package handlers

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/lixiang4u/imago/models"
	"github.com/lixiang4u/imago/utils"
	"path/filepath"
	"strings"
	"time"
)

func Process(ctx *fiber.Ctx) error {
	var userId = utils.TryParseUserId(ctx)
	var imgConfig = parseConfig(ctx)

	var file = filepath.Join(models.UploadRoot, "../", imgConfig.Src)
	if !strings.HasPrefix(utils.AbsPath(file), utils.AbsPath(models.UploadRoot)) {
		return ctx.JSON(respError("图片不存在"))
	}
	var fileSize = utils.FileSize(file)
	if fileSize <= 0 {
		return ctx.JSON(respError("图片不存在"))
	}

	var exportConfig = models.ExportConfig{
		StripMetadata: true,
		Quality:       int(imgConfig.Quality),
		Lossless:      false,
		Compression:   9,
	}
	appConfig, err := models.GetHostUserConfig(string(ctx.Request().Host()), userId)
	if err != nil {
		return ctx.JSON(respErrorDebug("系统错误", err.Error()))
	}

	var localMeta = models.LocalMeta{
		Id:          utils.FormattedUUID(16),
		FeatureId:   "default",
		Origin:      "",
		Remote:      false,
		Ext:         utils.ParseFileExt(file),
		Raw:         file,
		RemoteLocal: file,
		RequestUri:  imgConfig.Src,
		Size:        fileSize,
	}
	if !utils.IsDefaultObj(imgConfig, []string{"HttpAccept", "HttpUA", "Src"}) {
		localMeta.FeatureId = utils.HashString(fmt.Sprintf("%v", imgConfig))[:6]
	}

	var dstFormat = utils.GetFileMIME(localMeta.Raw).Subtype
	if len(imgConfig.Format) > 0 {
		dstFormat = imgConfig.Format
	}

	var convertedFile = fmt.Sprintf("%s.%s.%s", localMeta.Raw, localMeta.FeatureId, dstFormat)

	var requestLog = &models.RequestLog{
		UserId:     appConfig.UserId,
		ProxyId:    appConfig.ProxyId,
		MetaId:     localMeta.Id,
		RequestUrl: localMeta.RequestUri,
		OriginUrl:  localMeta.Raw,
		Referer:    ctx.Get("Referer"),
		UA:         imgConfig.HttpUA,
		Ip:         utils.GetClientIp(ctx),
		IsCache:    0,
		CreatedAt:  time.Now(),
	}

	if utils.FileExists(convertedFile) {
		var _size = utils.FileSize(convertedFile)

		go prepareShrinkLog(convertedFile, _size, 1, requestLog, &localMeta, &appConfig)

		return ctx.JSON(respSuccess(fiber.Map{
			"url":  fmt.Sprintf("%s/%s/%s", strings.TrimRight(appConfig.OriginSite, "/"), strings.Trim(models.FAKE_FILE_PREFIX, "/"), strings.TrimLeft(convertedFile, "/")),
			"path": fmt.Sprintf("/%s/%s", strings.Trim(models.FAKE_FILE_PREFIX, "/"), strings.TrimLeft(convertedFile, "/")),
			"size": _size,
			"rate": utils.CompressRate(fileSize, _size),
		}))
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
		return ctx.JSON(respError("图片处理失败：" + err.Error()))
	}

	go prepareShrinkLog(convertedFile, _size, 0, requestLog, &localMeta, &appConfig)

	return ctx.JSON(respSuccess(fiber.Map{
		"url":  fmt.Sprintf("%s/%s/%s", strings.TrimRight(appConfig.OriginSite, "/"), strings.Trim(models.FAKE_FILE_PREFIX, "/"), strings.TrimLeft(_converted, "/")),
		"path": fmt.Sprintf("/%s/%s", strings.Trim(models.FAKE_FILE_PREFIX, "/"), strings.TrimLeft(_converted, "/")),
		"size": _size,
		"rate": utils.CompressRate(fileSize, _size),
	}))
}
