package handlers

import (
	"github.com/davidbyttow/govips/v2/vips"
	"github.com/gofiber/fiber/v2"
	"github.com/lixiang4u/imago/models"
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

func parseConfig(ctx *fiber.Ctx) models.ImageConfig {
	var config = models.ImageConfig{}

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
	var appConfig = models.AppConfig{}

	// 域名映射：https://image.google.com/u/avatar.png => https://image-google.imago.com/u/avatar.png
	var p = vips.NewImportParams()
	p.FailOnError.Set(true)
	p.AutoRotate.Set(true)

	img, err := vips.LoadImageFromFile("/apps/repo/imago/images/cbh.png", p)
	if err != nil {
		log.Println("[err]", err.Error())
		return err
	}
	defer img.Close()

	img = ImageFilter(img, imgConfig, appConfig)

	buf, meta, err := ExportImage(img, img.Format(), models.ExportConfig{
		StripMetadata: true,
		Quality:       80,
		Lossless:      false,
	})
	if err != nil {
		log.Println("[err]", err.Error())
		return err
	}
	log.Println("[meta.Format]", meta.Format)

	if err = os.WriteFile("output-11.jpg", buf, 0600); err != nil {
		return err
	}

	return ctx.JSON(fiber.Map{
		"request":  strings.Split(ctx.Request().String(), "\r\n"),
		"Queries":  ctx.Queries(),
		"Hostname": ctx.Hostname(),
	})
}
