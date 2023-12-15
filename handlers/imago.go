package handlers

import (
	"github.com/davidbyttow/govips/v2/vips"
	"github.com/gofiber/fiber/v2"
	"log"
	"os"
	"runtime"
	"strings"
)

func init() {
	vips.Startup(&vips.Config{
		ConcurrencyLevel: runtime.NumCPU(),
	})

}

func Image(ctx *fiber.Ctx) error {

	// 域名映射：https://image.google.com/u/avatar.png => https://image-google.imago.com/u/avatar.png

	// 当前任务相关配置
	type Config struct {
		//StripMetadata：表示是否去除图片的元数据（如Exif信息等）。
		//Quality：表示图片的质量，取值范围为 1 到 100。较高的值表示更好的质量，但文件大小也会增加。
		//Lossless：表示是否使用无损压缩。如果设置为 true，则不会对图片进行任何有损压缩，保持完全的像素精度；如果设置为 false，则会应用有损压缩以减小文件大小。
		//NearLossless：表示是否使用近似无损压缩。如果设置为 true，则会应用一定程度的有损压缩，但保持较高的视觉质量；如果设置为 false，则不会应用近似无损压缩。
		//ReductionEffort：表示压缩的努力程度。取值范围为 0 到 6，0 表示最快压缩速度但压缩率较低，6 表示最慢压缩速度但压缩率较高。
		Quality         int
		Lossless        bool
		NearLossless    bool
		ReductionEffort int
	}
	var conf = Config{
		Quality:         80,
		Lossless:        false,
		NearLossless:    true,
		ReductionEffort: 0,
	}

	var p = vips.NewImportParams()
	p.FailOnError.Set(true)
	p.AutoRotate.Set(false)

	img, err := vips.LoadImageFromFile("./images/cbh.png", p)
	if err != nil {
		log.Println("[err]", err.Error())
		return err
	}
	defer img.Close()

	//options := webp.Options{Lossless: false}
	var buf []byte

	switch img.Format() {
	case vips.ImageTypeAVIF:
		fallthrough
	case vips.ImageTypePNG:
		fallthrough
	case vips.ImageTypeBMP:
		fallthrough
	case vips.ImageTypeJPEG:
		fallthrough
	default:
		// If some special images cannot encode with default ReductionEffort(0), then retry from 0 to 6
		var ep = vips.WebpExportParams{
			StripMetadata:   true,
			Quality:         conf.Quality,
			Lossless:        conf.Lossless,
			NearLossless:    conf.NearLossless,
			ReductionEffort: conf.ReductionEffort,
		}
		buf, _, err = img.ExportWebp(&ep)
	}

	if err = os.WriteFile("output.jpg", buf, 0600); err != nil {
		return err
	}

	return ctx.JSON(fiber.Map{
		"request":  strings.Split(ctx.Request().String(), "\r\n"),
		"Queries":  ctx.Queries(),
		"Hostname": ctx.Hostname(),
	})
}
