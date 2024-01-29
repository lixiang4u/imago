package handlers

import (
	"fmt"
	"github.com/davidbyttow/govips/v2/vips"
	"github.com/gofiber/fiber/v2"
	"github.com/lixiang4u/imago/models"
	"github.com/lixiang4u/imago/utils"
	"log"
	"os"
	"path"
	"runtime"
	"sync"
	"time"
)

func init() {
	vips.Startup(&vips.Config{
		ConcurrencyLevel: runtime.NumCPU(),
	})

}

func parseConfig(ctx *fiber.Ctx) models.ImageConfig {
	var config = models.ImageConfig{}

	switch ctx.Method() {
	case "POST":
		_ = ctx.BodyParser(&config)
	case "GET":
		fallthrough
	default:
		_ = ctx.QueryParser(&config)
	}

	log.Println("[QueryParser]", ctx.OriginalURL(), utils.ToJsonString(config, false))

	config.HttpUA = string(ctx.Request().Header.Peek("User-Agent"))
	config.HttpAccept = string(ctx.Request().Header.Peek("Accept"))

	return config
}

func Image(ctx *fiber.Ctx) error {
	var now = time.Now()
	var imgConfig = parseConfig(ctx)
	var appConfig models.AppConfig
	appConfig, err := models.GetHostUserConfig(string(ctx.Request().Host()))
	if err != nil {
		return ctx.JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	localMeta, err := HandleToLocalPath(ctx, &imgConfig, &appConfig)
	if err != nil {
		return ctx.JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	var supported = CheckSupported(imgConfig.HttpAccept, imgConfig.HttpUA)

	var exportConfig = models.ExportConfig{
		StripMetadata: true,
		Quality:       int(imgConfig.Quality),
		Lossless:      false,
	}

	var requestLog = &models.RequestLog{
		UserId:     appConfig.UserId,
		ProxyId:    appConfig.ProxyId,
		MetaId:     localMeta.Id,
		RequestUrl: localMeta.RequestUri,
		OriginUrl:  localMeta.Raw,
		Referer:    ctx.Get("Referer"),
		UA:         imgConfig.HttpUA,
		Ip:         utils.GetClientIp(ctx),
		IsCache:    1,
		CreatedAt:  now,
	}
	if localMeta.FetchSource {
		requestLog.IsCache = 0
	}

	reqCount, _ := models.IncrementRequestCount(&appConfig)

	// 源文件不存在
	if !utils.FileExists(localMeta.RemoteLocal) {
		go func() { _ = prepareRequestLog(requestLog, 0) }()
		//utils.RemoveMeta(localMeta.Id, localMeta.Origin)
		_ = ctx.Send([]byte("raw file not found"))
		_ = ctx.SendStatus(404)
		return nil
	}

	convertedFile, convertedSize, ok := ConvertAndGetSmallestImage(localMeta, supported, &imgConfig, &exportConfig)
	if !ok {
		_ = ctx.Send([]byte("convert failed"))
		_ = ctx.SendStatus(404)
		return nil
	}

	var mime = utils.GetFileMIME(convertedFile)
	reqOkCount, _ := models.IncrementRequestOkCount(&appConfig)

	ctx.Set("Content-Type", mime.Value)
	if len(appConfig.Cors) > 0 {
		ctx.Set("Access-Control-Allow-Origin", appConfig.Cors)
	}
	ctx.Set("X-Compression-Rate", utils.CompressRate(localMeta.Size, convertedSize))
	ctx.Set("X-Server", "imago")
	if appConfig.Debug {
		ctx.Set("X-Stat", fmt.Sprintf("total: %d, success:%d", reqCount, reqOkCount))
	}

	// 统计
	go prepareShrinkLog(convertedFile, convertedSize, 1, requestLog, &localMeta, &appConfig)

	return ctx.SendFile(convertedFile)
}

func ConvertAndGetSmallestImage(
	localMeta models.LocalMeta,
	supported map[string]bool,
	imgConfig *models.ImageConfig,
	exportConfig *models.ExportConfig,
) (convertedFile string, size int64, ok bool) {
	// 默认源文件最小，但是也可能压缩后的问题比源文件还大（源文件是压缩文件会导致再次压缩会变大）
	size = localMeta.Size
	convertedFile = localMeta.RemoteLocal

	var wg sync.WaitGroup
	for fileType, ok := range supported {
		if !ok {
			continue
		}
		switch fileType {
		case models.SUPPORT_TYPE_RAW:
			fileType = localMeta.Ext
			fallthrough
		case models.SUPPORT_TYPE_WEBP:
			fallthrough
		case models.SUPPORT_TYPE_AVIF:
			fallthrough
		case models.SUPPORT_TYPE_JPG:
			fallthrough
		default:
			wg.Add(1)
			go func(rawFile, fileType string) {
				defer wg.Done()
				_converted, _size, err := ConvertImage(
					rawFile,
					fmt.Sprintf(
						"%s.%s.%s",
						utils.GetOutputFilePath(localMeta.Id, localMeta.Origin, localMeta.Ext),
						localMeta.FeatureId,
						fileType,
					),
					fileType,
					imgConfig,
					exportConfig,
				)
				if err != nil {
					log.Println("[convert image]", err.Error())
					return
				}
				if size == 0 || _size < size {
					size = _size
					convertedFile = _converted
				}
			}(localMeta.RemoteLocal, fileType)
		}
	}
	wg.Wait()

	return convertedFile, size, len(convertedFile) > 0
}

func ConvertImage(
	rawFile,
	convertedFile,
	format string,
	imgConfig *models.ImageConfig,
	exportConfig *models.ExportConfig,
) (converted string, size int64, err error) {
	converted = convertedFile
	if utils.FileExists(convertedFile) {
		return converted, utils.FileSize(convertedFile), nil
	}

	if err = os.MkdirAll(path.Dir(convertedFile), 0666); err != nil {
		log.Println("[convert mkdir]", path.Dir(convertedFile), err.Error())
		return
	}

	var p = vips.NewImportParams()
	p.FailOnError.Set(true)
	if imgConfig.XAutoRotate {
		p.AutoRotate.Set(true)
	}

	img, err := vips.LoadImageFromFile(rawFile, p)
	if err != nil {
		log.Println("[libvips load]", err.Error())
		return converted, 0, err
	}
	defer img.Close()

	log.Println("[converting]", rawFile, "=>", convertedFile, img.Format().FileExt(), "=>", format, "config:", utils.ToJsonString(exportConfig, false))

	img = ImageFilter(img, imgConfig)

	buf, _, err := ExportImage(img, format, exportConfig)
	if err != nil {
		log.Println("[export image]", err.Error())
		return converted, 0, err
	}
	log.Println("[export image]", rawFile, "=>", convertedFile)

	if err := os.WriteFile(convertedFile, buf, 0600); err != nil {
		log.Println("[export save]", err.Error())
		return converted, 0, err
	}

	return converted, utils.FileSize(convertedFile), nil
}
