package handlers

import (
	"fmt"
	"github.com/davidbyttow/govips/v2/vips"
	"github.com/gofiber/fiber/v2"
	"github.com/lixiang4u/imago/models"
	"github.com/lixiang4u/imago/utils"
	"log"
	"os"
	"runtime"
	"sync"
)

func init() {
	vips.Startup(&vips.Config{
		ConcurrencyLevel: runtime.NumCPU(),
	})

}

func parseConfig(ctx *fiber.Ctx) models.ImageConfig {
	var config = models.ImageConfig{}

	//	config.HttpAccept = ctx.Query("Accept")
	//	config.HttpUA = ctx.Query("User-Agent")

	// 特殊字段，指定图片源地址，本地/网络地址
	config.Src = ctx.Query("src")

	//height //width //both are in px.
	config.With = ctx.QueryFloat("width")
	config.Height = ctx.QueryFloat("height")

	// Available options are: v(vertical), h(horizontal), b(Both vertical and horizontal)
	config.Flip = ctx.Query("flip")

	// Override quality set in dashbaord, available quality range from 10 ~ 100(100 means lossless convert)
	config.Quality = ctx.QueryFloat("quality")

	// Available blur range from 10 ~ 100
	config.Blur = ctx.QueryFloat("blur")

	// Sharpen the image, available sharpen range from 1 ~ 10
	config.Sharpen = ctx.QueryFloat("sharpen")

	// Available rotate angle range from 0 ~ 360, however if angle is not 90, 180, 270, 360, it will be filled with white background
	config.Rotate = ctx.QueryFloat("rotate")

	// Adjust brightness of the image, available range from 0 ~ 10, 1 means no change
	config.Brightness = ctx.QueryFloat("brightness")

	// Adjust saturation of the image, available range from 0 ~ 10, 1 means no change
	config.Saturation = ctx.QueryFloat("saturation")

	// Adjust hue of the image, available range from 0 ~ 360, hue will be 0 for no change, 90 for a complementary hue shift, 180 for a contrasting shift, 360 for no change again.
	config.Hue = ctx.QueryFloat("hue")

	// Adjust contrast of the image, available range from 0 ~ 10, 1 means no change
	config.Contrast = ctx.QueryFloat("contrast")

	//
	config.Watermark = struct {
		Text    string
		Font    string
		Color   string
		With    float64
		Height  float64
		OffsetX float64
		OffsetY float64
		Opacity float64
	}{}

	config.Watermark.Text = ctx.Query("text")

	config.Filter = ctx.Query("filter")

	return config
}

func Image(ctx *fiber.Ctx) error {
	var imgConfig = parseConfig(ctx)
	var appConfig = models.AppConfig{
		OriginSite: "",
		//LocalPath:  "",
	}
	localMeta, err := HandleToLocalPath(ctx, &appConfig)
	if err != nil {
		return ctx.JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	var supported = CheckSupported(imgConfig.HttpAccept, imgConfig.HttpUA)

	var exportConfig = models.ExportConfig{
		StripMetadata: true,
		Quality:       80,
		Lossless:      false,
	}

	log.Println("[rawLocalFile]", utils.ToJsonString(localMeta))

	log.Println("[drop]", utils.HashString(utils.ToJsonString(fiber.Map{
		"imgConfig":    imgConfig,
		"appConfig":    appConfig,
		"exportConfig": exportConfig,
	})))

	// 源文件不存在
	if !utils.FileExists(localMeta.RemoteLocal) {
		utils.RemoveMeta(localMeta.Id, localMeta.Origin)
		_ = ctx.Send([]byte("file not found"))
		_ = ctx.SendStatus(404)
		return nil
	}

	convertedFile, convertedSize, ok := ConvertAndGetSmallestImage(localMeta.RemoteLocal, supported, &imgConfig, &appConfig, &exportConfig)
	if !ok {
		_ = ctx.Send([]byte("convert failed"))
		_ = ctx.SendStatus(404)
		return nil
	}

	var mime = utils.GetFileMIME(convertedFile)

	log.Println("[debug]", utils.ToJsonString(fiber.Map{
		"localMeta.RemoteLocal": localMeta.RemoteLocal,
		"convertedFile":         convertedFile,
		"localMeta.Size":        localMeta.Size,
		"convertedSize":         convertedSize,
		"mime":                  mime,
	}))

	ctx.Set("Content-Type", mime.Value)
	ctx.Set("X-Compression-Rate", utils.CompressRate(localMeta.Size, convertedSize))
	return ctx.SendFile(convertedFile)
}

func ConvertAndGetSmallestImage(
	rawFile string,
	supported map[string]bool,
	imgConfig *models.ImageConfig,
	appConfig *models.AppConfig,
	exportConfig *models.ExportConfig,
) (convertedFile string, size int64, ok bool) {
	var wg sync.WaitGroup
	for _, ok := range supported {
		if !ok {
			continue
		}
		wg.Add(1)
	}

	for fileType, ok := range supported {
		if !ok {
			continue
		}
		switch fileType {
		case models.SUPPORT_TYPE_WEBP:
			fallthrough
		case models.SUPPORT_TYPE_AVIF:
			fallthrough
		case models.SUPPORT_TYPE_JPG:
			go func() {
				defer wg.Done()
				_converted, _size, err := ConvertImage(rawFile, fmt.Sprintf("%s.c.%s", rawFile, models.SUPPORT_TYPE_JPG), models.SUPPORT_TYPE_JPG, imgConfig, appConfig, exportConfig)
				if err != nil {
					log.Println("[ConvertImageError]", err.Error())
					return
				}
				if size == 0 || _size < size {
					size = _size
					convertedFile = _converted
				}
			}()
		default:
			break
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
	appConfig *models.AppConfig,
	exportConfig *models.ExportConfig,
) (converted string, size int64, err error) {
	converted = convertedFile
	if utils.FileExists(convertedFile) {
		return converted, utils.FileSize(convertedFile), nil
	}

	var p = vips.NewImportParams()
	p.FailOnError.Set(true)
	p.AutoRotate.Set(true)

	img, err := vips.LoadImageFromFile(rawFile, p)
	if err != nil {
		log.Println("[vips.LoadImageFromFile]", err.Error())
		return converted, 0, err
	}
	defer img.Close()

	img = ImageFilter(img, imgConfig, appConfig)

	buf, meta, err := ExportImage(img, format, exportConfig)
	if err != nil {
		log.Println("[ExportImage]", err.Error())
		return converted, 0, err
	}
	log.Println("[ExportImage.Format]", meta.Format)

	if err := os.WriteFile(convertedFile, buf, 0600); err != nil {
		log.Println("[ExportImageSave]", err.Error())
		return converted, 0, err
	}

	return converted, utils.FileSize(convertedFile), nil
}
