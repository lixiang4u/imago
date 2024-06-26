package handlers

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/davidbyttow/govips/v2/vips"
	"github.com/gofiber/fiber/v2"
	"github.com/lixiang4u/imago/models"
	"github.com/lixiang4u/imago/utils"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"slices"
	"strings"
)

func downloadFile(remotePath, localPath string, appConfig *models.AppConfig) error {
	req, err := http.NewRequest("GET", remotePath, nil)
	if err != nil {
		log.Println("[download error0]", remotePath, err.Error())
		return err
	}
	req.Header.Set("User-Agent", appConfig.UserAgent)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("[download error1]", remotePath, err.Error())
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		log.Println("[download error2]", remotePath, resp.StatusCode)
		return errors.New(fmt.Sprintf("fetch source status: %d", resp.StatusCode))
	}
	if resp.ContentLength > int64(models.MaxUpload) {
		log.Println("[download error5]", remotePath, resp.ContentLength)
		return errors.New(fmt.Sprintf("source too big: %d", resp.ContentLength))
	}
	// 下载文件
	if err = os.MkdirAll(path.Dir(localPath), 0666); err != nil {
		log.Println("[download error3]", path.Dir(localPath), err.Error())
		return err
	}

	models.LocalCache.Set(localPath, true, -1)
	defer func() {
		models.LocalCache.Delete(localPath)
	}()

	var buf = bytes.Buffer{}
	_, _ = buf.ReadFrom(resp.Body)
	if err = os.WriteFile(localPath, buf.Bytes(), 0666); err != nil {
		log.Println("[download error4]", localPath, err.Error())
		return err
	}

	return nil
}

func CheckFileAllowed(fileName string) (pathUri, ext string, ok bool) {
	u, _ := url.Parse(fileName)
	ext = strings.ToLower(strings.TrimPrefix(path.Ext(u.Path), "."))
	if slices.Contains(models.ImageTypes, ext) {
		return u.Path, ext, true
	}
	return u.Path, ext, false
}

func CheckSupported(httpAccept, httpUA string) map[string]bool {
	var supported = models.LocalConfig.L.Supported

	if strings.Contains(httpAccept, "image/webp") {
		supported[models.SUPPORT_TYPE_WEBP] = true
	}
	// lib暂不支持
	//if strings.Contains(httpAccept, "image/avif") {
	//	supported[models.SUPPORT_TYPE_AVIF] = true
	//}
	if strings.Contains(httpAccept, "image/jpeg") {
		supported[models.SUPPORT_TYPE_JPG] = true
	}
	if strings.Contains(httpAccept, "image/jpg") {
		supported[models.SUPPORT_TYPE_JPG] = true
	}
	if strings.Contains(httpAccept, "image/pjpeg") {
		supported[models.SUPPORT_TYPE_JPG] = true
	}
	if strings.Contains(httpAccept, "*/*") {
		for k, _ := range supported {
			supported[k] = true
		}
	}

	return supported
}

func HandleLocalMeta(pathOrUri string, imgConfig *models.ImageConfig, appConfig *models.AppConfig) (models.LocalMeta, error) {
	var rawFileClean = "" // 不带参数的路径
	var localMeta = models.LocalMeta{
		Id:         "",
		FeatureId:  "default",
		RequestUri: pathOrUri,
		Ext:        "raw",
		RawVersion: "",
		Size:       0,
		//Origin:      "",
		//Remote:      false,
		//RemoteLocal: "",
		//Raw:         "",
	}

	tmpUrl, err := url.Parse(pathOrUri)
	if err != nil {
		rawFileClean = pathOrUri
	} else {
		rawFileClean = tmpUrl.Path
	}

	if len(appConfig.OriginSite) != 0 {
		localMeta.Remote = true
		localMeta.Origin = utils.ParseUrlDefaultHost(appConfig.OriginSite, models.Local)
		localMeta.Raw = fmt.Sprintf("%s/%s", strings.TrimRight(appConfig.OriginSite, "/"), strings.TrimLeft(pathOrUri, "/"))
	} else {
		// 不能使用带参数的uri路径
		localMeta.Remote = false
		localMeta.Origin = models.Local
		localMeta.Raw = fmt.Sprintf("%s/%s", strings.TrimRight(appConfig.LocalPath, "/"), strings.TrimLeft(rawFileClean, "/"))
	}

	// 请求路径能解析出文件后缀
	var ext = strings.ToLower(strings.TrimPrefix(path.Ext(rawFileClean), "."))
	if slices.Contains(models.ImageTypes, ext) {
		localMeta.Ext = ext
	}

	localMeta.Id = utils.HashString(fmt.Sprintf("%s,%s", appConfig.OriginSite, rawFileClean))
	if !utils.IsDefaultObj(*imgConfig, []string{"HttpAccept", "HttpUA", "Src"}) {
		localMeta.FeatureId = utils.HashString(fmt.Sprintf("%v", imgConfig))[:6]
	}

	if localMeta.Remote {
		localMeta.RemoteLocal = utils.GetRemoteLocalFilePath(localMeta.Id, localMeta.Origin, localMeta.Ext)
	} else {
		localMeta.RemoteLocal = rawFileClean
	}

	return localMeta, nil
}

func HandleToLocalPath(ctx *fiber.Ctx, imgConfig *models.ImageConfig, appConfig *models.AppConfig) (models.LocalMeta, error) {
	localMeta, err := HandleLocalMeta(string(ctx.Request().RequestURI()), imgConfig, appConfig)
	if err != nil {
		return localMeta, err
	}

	meta, err := utils.GetMeta(localMeta.Id, localMeta.Origin, localMeta.Raw, localMeta.RawVersion)
	if err != nil {
		return localMeta, err
	}

	localMeta.Size = utils.FileSize(localMeta.RemoteLocal)

	// 本地文件
	if !localMeta.Remote {
		return localMeta, nil
	}

	var rawExists = utils.FileExists(localMeta.RemoteLocal)
	// 如果不需要refresh且文件存在，直接返回
	if rawExists && appConfig.Refresh == 0 {
		return localMeta, nil
	}
	// 如果需要refresh且版本未变化，直接返回
	if rawExists && appConfig.Refresh != 0 && meta.Version == localMeta.RawVersion {
		return localMeta, nil
	}

	// 需要回源，清理老数据
	utils.RemoveCache(localMeta.RemoteLocal)
	utils.RemoveMeta(localMeta.Id, localMeta.Origin)
	utils.LogMeta(localMeta.Id, localMeta.Origin, localMeta.Raw, localMeta.RawVersion)

	log.Println("[fetchSource]", localMeta.Raw, "=>", localMeta.RemoteLocal)

	if err = downloadFile(localMeta.Raw, localMeta.RemoteLocal, appConfig); err != nil {
		return localMeta, err
	}

	localMeta.Size = utils.FileSize(localMeta.RemoteLocal)
	localMeta.FetchSource = true

	return localMeta, nil
}

func ImageFilter(img *vips.ImageRef, imgConfig *models.ImageConfig) *vips.ImageRef {
	_ = _filter(img, imgConfig)
	return img
}

func _filter(img *vips.ImageRef, imgConfig *models.ImageConfig) (err error) {
	if img.Metadata().Width > models.MaxWebpPixel {
		img.Metadata().Width = models.MaxWebpPixel
	}
	if img.Metadata().Height > models.MaxWebpPixel {
		img.Metadata().Height = models.MaxWebpPixel
	}

	err = nil
	var imgRatio = float32(img.Metadata().Width) / float32(img.Metadata().Height)
	if imgConfig.Width > 0 && imgConfig.Height > 0 {
		err = img.Thumbnail(int(imgConfig.Width), int(imgConfig.Height), vips.InterestingAttention)
	} else if imgConfig.Width > 0 && imgConfig.Height == 0 {
		err = img.Thumbnail(int(imgConfig.Width), int(float32(imgConfig.Width)/imgRatio), 0)
	} else if imgConfig.Height > 0 && imgConfig.Width == 0 {
		err = img.Thumbnail(int(float32(imgConfig.Height)*imgRatio), int(imgConfig.Height), 0)
	}
	if err != nil {
		log.Println("[image resize]", err.Error())
	}

	err = nil
	switch strings.ToLower(imgConfig.Flip) {
	case "v":
		fallthrough
	case "vertical":
		err = img.Flip(vips.DirectionVertical)
	case "h":
		fallthrough
	case "horizontal":
		err = img.Flip(vips.DirectionHorizontal)
	case "b":
		fallthrough
	case "both":
		err = img.Flip(vips.DirectionHorizontal | vips.DirectionVertical)
	}
	if err != nil {
		log.Println("[image flip]", err.Error())
	}

	err = nil
	if imgConfig.Blur > 0 {
		err = img.GaussianBlur(imgConfig.Blur)
	}
	if err != nil {
		log.Println("[image blur]", err.Error())
	}

	err = nil
	if imgConfig.Sharpen > 0 {
		err = img.Sharpen(imgConfig.Sharpen, 0, 0)
	}
	if err != nil {
		log.Println("[image sharpen]", err.Error())
	}

	err = nil
	if imgConfig.Rotate > 0 {
		switch imgConfig.Rotate {
		case 0:
			err = img.Rotate(vips.Angle0)
		case 90:
			err = img.Rotate(vips.Angle90)
		case 180:
			err = img.Rotate(vips.Angle180)
		case 270:
			err = img.Rotate(vips.Angle270)
		case 360:
			err = img.Rotate(vips.Angle0)
		}
	}
	if err != nil {
		log.Println("[image rotate]", err.Error())
	}

	err = nil
	if imgConfig.Brightness > 0 || imgConfig.Saturation > 0 || imgConfig.Hue > 0 {
		err = img.Modulate(imgConfig.Brightness, imgConfig.Saturation, imgConfig.Hue)
	}
	if err != nil {
		log.Println("[image modulate]", err.Error())
	}

	// contrast 暂不支持
	//img.Label(&vips.LabelParams{
	//	Text:      "",
	//	Font:      "",
	//	Width:     vips.Scalar{},
	//	Height:    vips.Scalar{},
	//	OffsetX:   vips.Scalar{},
	//	OffsetY:   vips.Scalar{},
	//	Opacity:   0,
	//	Color:     vips.Color{},
	//	Alignment: 0,
	//})

	return err
}

func ExportImage(img *vips.ImageRef, toType string, exportParams *models.ExportConfig) (buf []byte, meta *vips.ImageMetadata, err error) {
	switch toType {
	case models.SUPPORT_TYPE_RAW:
		buf, meta, err = img.ExportNative()
	case models.SUPPORT_TYPE_WEBP:
		// If some special images cannot encode with default ReductionEffort(0), then retry from 0 to 6
		buf, meta, err = img.ExportWebp(&vips.WebpExportParams{
			StripMetadata:   exportParams.StripMetadata,
			Lossless:        exportParams.Lossless,
			Quality:         exportParams.Quality,
			ReductionEffort: exportParams.ReductionEffort,
		})
	case models.SUPPORT_TYPE_AVIF:
		buf, meta, err = img.ExportAvif(&vips.AvifExportParams{
			StripMetadata: true,
			Quality:       exportParams.Quality,
			Lossless:      exportParams.Lossless,
		})
	case models.SUPPORT_TYPE_BMP:
		fallthrough
	case models.SUPPORT_TYPE_JPEG:
		fallthrough
	case models.SUPPORT_TYPE_JPG:
		buf, meta, err = img.ExportJpeg(&vips.JpegExportParams{
			StripMetadata:  true,
			Quality:        exportParams.Quality,
			OptimizeCoding: exportParams.OptimizeCoding,
		})
	case models.SUPPORT_TYPE_PNG:
		buf, meta, err = img.ExportPng(&vips.PngExportParams{
			StripMetadata: true,
			Filter:        vips.PngFilterAll,
			Compression:   exportParams.Compression,
			Quality:       exportParams.Quality,
		})
	case models.SUPPORT_TYPE_GIF:
		buf, meta, err = img.ExportGIF(&vips.GifExportParams{
			StripMetadata: true,
			Quality:       exportParams.Quality,
		})
	default:
		buf, meta, err = img.ExportNative()
	}
	return
}

//func ImageFilter(img *vips.ImageRef, config *models.ImageConfig, appConfig *models.AppConfig) *vips.ImageRef {
//	return img
//}
//func ExportImage(img *vips.ImageRef, toType string, exportParams *models.ExportConfig) (buf []byte, meta *vips.ImageMetadata, err error) {
//	return
//}
