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

func downloadFile(remotePath, localPath string) error {
	resp, err := http.Get(remotePath)
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

func CheckFileAllowed(fileName string) (ext string, ok bool) {
	u, _ := url.Parse(fileName)
	ext = strings.ToLower(strings.TrimPrefix(path.Ext(u.Path), "."))
	if slices.Contains(models.ImageTypes, ext) {
		return ext, true
	}
	return ext, false
}

func CheckSupported(httpAccept, httpUA string) map[string]bool {
	var supported = map[string]bool{
		models.SUPPORT_TYPE_RAW:  true,
		models.SUPPORT_TYPE_WEBP: false,
		models.SUPPORT_TYPE_AVIF: false,
		models.SUPPORT_TYPE_JPG:  false,
	}

	if strings.Contains(httpAccept, "image/webp") {
		supported["webp"] = true
	}
	if strings.Contains(httpAccept, "image/avif") {
		supported["avif"] = true
	}
	if strings.Contains(httpAccept, "image/jpeg") {
		supported["jpg"] = true
	}
	if strings.Contains(httpAccept, "image/jpg") {
		supported["jpg"] = true
	}
	if strings.Contains(httpAccept, "image/pjpeg") {
		supported["jpg"] = true
	}
	if strings.Contains(httpAccept, "*/*") {
		for k, _ := range supported {
			supported[k] = true
		}
	}

	return supported
}

func HandleToLocalPath(ctx *fiber.Ctx, imgConfig *models.ImageConfig, appConfig *models.AppConfig) (models.LocalMeta, error) {
	var remote = false
	var rawFile = models.Empty
	var originHost = models.Local
	var rawVersion = models.Empty
	var requestUri = string(ctx.Request().RequestURI())
	var localMeta models.LocalMeta

	fileExt, ok := CheckFileAllowed(requestUri)
	if !ok {
		return localMeta, errors.New("file type not support")
	}

	if len(appConfig.OriginSite) != 0 {
		remote = true
		tmpUrl, err := url.Parse(appConfig.OriginSite)
		if err != nil {
			tmpUrl.Host = models.Local
		}
		originHost = tmpUrl.Host
		rawFile = fmt.Sprintf("%s/%s", strings.TrimRight(appConfig.OriginSite, "/"), strings.TrimLeft(requestUri, "/"))
		if appConfig.Refresh == 1 {
			rawVersion = utils.GetResourceVersion(rawFile, nil)
		}
	} else {
		rawFile = path.Join(appConfig.LocalPath, requestUri)
	}

	var id = utils.HashString(fmt.Sprintf("%s,%s", rawFile, appConfig.OriginSite))
	var featureId = "default"
	if !utils.IsDefaultObj(imgConfig) {
		featureId = utils.HashString(fmt.Sprintf("%v", imgConfig))[:6]
	}

	localMeta = models.LocalMeta{
		Id:          id,
		FeatureId:   featureId,
		Remote:      remote,
		Origin:      originHost,
		Ext:         fileExt,
		RemoteLocal: rawFile,
		Raw:         rawFile,
		Size:        0,
	}

	meta, err := utils.GetMeta(id, originHost, rawFile, rawVersion)
	if err != nil {
		return localMeta, err
	}

	if !remote {
		localMeta.Size = utils.FileSize(rawFile)
		return localMeta, nil
	}

	localMeta.RemoteLocal = utils.GetRemoteLocalFilePath(id, originHost, fileExt)

	var rawExists = utils.FileExists(localMeta.RemoteLocal)
	// 如果不需要refresh且文件存在，直接返回
	if rawExists && appConfig.Refresh == 0 {
		localMeta.Size = utils.FileSize(localMeta.RemoteLocal)
		return localMeta, nil
	}
	// 如果需要refresh且版本未变化，直接返回
	if rawExists && appConfig.Refresh != 0 && meta.Version == rawVersion {
		localMeta.Size = utils.FileSize(localMeta.RemoteLocal)
		return localMeta, nil
	}

	// 需要回源，清理老数据
	utils.RemoveCache(localMeta.RemoteLocal)
	utils.RemoveMeta(id, originHost)
	utils.LogMeta(id, originHost, rawFile, rawVersion)

	log.Println("[fetch source]", rawFile, "=>", localMeta.RemoteLocal)

	if err = downloadFile(rawFile, localMeta.RemoteLocal); err != nil {
		return localMeta, err
	}

	localMeta.Size = utils.FileSize(localMeta.RemoteLocal)

	return localMeta, nil
}

func ImageFilter(img *vips.ImageRef, imgConfig *models.ImageConfig, appConfig *models.AppConfig) *vips.ImageRef {
	_ = _filter(img, imgConfig)
	return img
}

func _filter(img *vips.ImageRef, imgConfig *models.ImageConfig) (err error) {
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
		err = img.Rotate(vips.Angle(int(imgConfig.Rotate)))
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
		buf, meta, err = img.ExportAvif(vips.NewAvifExportParams())
		fallthrough
	case models.SUPPORT_TYPE_JPG:
		buf, meta, err = img.ExportJpeg(vips.NewJpegExportParams())
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
