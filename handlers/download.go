package handlers

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/lixiang4u/imago/models"
	"github.com/lixiang4u/imago/utils"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

// Download https://www.imago.xyz/api/file/upload/d9157.png?filename=1.png&download
func Download(ctx *fiber.Ctx) error {
	var isDownload = false
	var fileName = ctx.Query("filename", filepath.Base(ctx.Params("*")))
	if _, ok := ctx.Queries()["download"]; ok {
		isDownload = true
	}
	uploadRoot, err := filepath.Abs(path.Dir(models.UploadRoot))
	if err != nil {
		return ctx.JSON(respErrorDebug("上传目录配置异常", err.Error()))
	}
	requestFile, err := filepath.Abs(path.Join(path.Dir(models.UploadRoot), ctx.Params("*")))
	if err != nil {
		return ctx.JSON(respErrorDebug("文件不存在", err.Error()))
	}
	if !strings.HasPrefix(requestFile, uploadRoot) {
		return ctx.JSON(respErrorDebug("下载文件地址异常", ctx.Params("*")))
	}
	if utils.FileSize(requestFile) <= 0 {
		return ctx.JSON(respErrorDebug("文件不存在", ctx.Params("*")))
	}
	if isDownload {
		return ctx.Download(requestFile, fileName)
	}
	return ctx.SendFile(requestFile)
}

func Archive(ctx *fiber.Ctx) error {
	var zipName = fmt.Sprintf("imago-%s.zip", time.Now().Format("20060102150405"))

	type Req struct {
		Files []models.SimpleFile `json:"files" form:"files"`
	}
	var req Req
	_ = ctx.BodyParser(&req)

	uploadRoot, err := filepath.Abs(path.Dir(models.UploadRoot))
	if err != nil {
		return ctx.JSON(respErrorDebug("上传目录配置异常", err.Error()))
	}

	var sourceFiles []models.SimpleFile
	for i, file := range req.Files {
		if i >= 100 {
			break
		}
		tmpFile, err := filepath.Abs(path.Join(uploadRoot, strings.TrimPrefix(file.Path, models.FAKE_FILE_PREFIX)))
		if err != nil {
			continue
		}
		if !strings.HasPrefix(tmpFile, uploadRoot) {
			continue
		}
		_, err = os.Stat(tmpFile)
		if err != nil {
			continue
		}
		file.Path = tmpFile
		sourceFiles = append(sourceFiles, file)
	}
	if len(sourceFiles) >= 100 {
		return ctx.JSON(respError("打包文件数过多"))
	}
	if len(sourceFiles) <= 0 {
		return ctx.JSON(respError("打包文件不存在"))
	}

	var zipFile = path.Join(os.TempDir(), fmt.Sprintf("imago_tmp_%d_%s.zip", time.Now().Unix(), utils.HashString(fmt.Sprintf("%d", time.Now().UnixNano()))[:8]))

	n, err := utils.CreateZip(zipFile, sourceFiles)
	if err != nil {
		return ctx.JSON(respErrorDebug("打包失败", err.Error()))
	}
	if n <= 0 {
		return ctx.JSON(respError("打包文件列表异常"))
	}
	defer func() { _ = os.Remove(zipFile) }()

	// https://developer.mozilla.org/zh-CN/docs/Glossary/Simple_response_header
	ctx.Response().Header.Add("X-zip_name", zipName)

	return ctx.Download(zipFile, zipName)
}
