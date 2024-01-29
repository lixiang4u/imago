package handlers

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/lixiang4u/imago/models"
	"github.com/lixiang4u/imago/utils"
	"log"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

func Process(ctx *fiber.Ctx) error {
	type PostRequest struct {
		Path string `json:"path" form:"path"`
	}

	var postRequest PostRequest
	_ = ctx.BodyParser(&postRequest)

	var file = filepath.Join(models.UploadRoot, "../", postRequest.Path)
	if !strings.HasPrefix(utils.AbsPath(file), utils.AbsPath(models.UploadRoot)) {
		return ctx.JSON(fiber.Map{
			"error": "图片不存在",
		})
	}
	var fileSize = utils.FileSize(file)
	if fileSize <= 0 {
		return ctx.JSON(fiber.Map{
			"error": "图片不存在",
		})
	}

	var userId = utils.TryParseUserId(ctx)
	var imgConfig = parseConfig(ctx)
	var exportConfig = models.ExportConfig{
		StripMetadata: true,
		Quality:       int(imgConfig.Quality),
		Lossless:      false,
		Compression:   9,
	}
	appConfig, err := models.GetHostUserConfig(string(ctx.Request().Host()), userId)
	if err != nil {
		return ctx.JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	log.Println("[appConfig]", utils.ToJsonString(appConfig, true))

	var localMeta = models.LocalMeta{
		Id:          utils.FormattedUUID(16),
		FeatureId:   "default",
		Origin:      "",
		Remote:      false,
		Ext:         utils.ParseFileExt(file),
		Raw:         file,
		RemoteLocal: file,
		RequestUri:  postRequest.Path,
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
		"url":    fmt.Sprintf("%s/%s/%s", strings.TrimRight(appConfig.OriginSite, "/"), strings.Trim(models.FAKE_FILE_PREFIX, "/"), strings.TrimLeft(_converted, "/")),
		"path":   fmt.Sprintf("/%s/%s", strings.Trim(models.FAKE_FILE_PREFIX, "/"), strings.TrimLeft(_converted, "/")),
		"size":   _size,
		"rate":   utils.CompressRate(fileSize, _size),
	})
}

func Shrink(ctx *fiber.Ctx) error {
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

	var imgConfig = parseConfig(ctx)
	var exportConfig = models.ExportConfig{
		StripMetadata: true,
		Quality:       int(imgConfig.Quality),
		Lossless:      false,
		Compression:   9,
	}
	appConfig, err := models.GetHostUserConfig(string(ctx.Request().Host()), userId)
	if err != nil {
		return ctx.JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	log.Println("[userId]", userId)
	log.Println("[appConfig]", utils.ToJsonString(appConfig, true))

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
		Id:         utils.FormattedUUID(16),
		FeatureId:  "default",
		Origin:     "",
		Remote:     false,
		Ext:        strings.ToLower(strings.Trim(path.Ext(fh.Filename), ".")),
		Raw:        "",
		RequestUri: string(ctx.Request().RequestURI()),
		Size:       fh.Size,
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
		"status":    "ok",
		"file_name": fh.Filename,
		"file_size": fh.Size,
		"url":       fmt.Sprintf("%s/%s/%s", strings.TrimRight(appConfig.OriginSite, "/"), strings.Trim(models.FAKE_FILE_PREFIX, "/"), strings.TrimLeft(_converted, "/")),
		"path":      fmt.Sprintf("/%s/%s", strings.Trim(models.FAKE_FILE_PREFIX, "/"), strings.TrimLeft(_converted, "/")),
		"size":      _size,
		"rate":      utils.CompressRate(fh.Size, _size),
		"time":      time.Now().Format("2006-01-02 15:04:05"),
	})
}
