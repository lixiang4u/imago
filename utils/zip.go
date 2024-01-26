package utils

import (
	"archive/zip"
	"github.com/lixiang4u/imago/models"
	"io"
	"log"
	"os"
)

func CreateZip(zipFile string, sourceFiles []models.SimpleFile) (n int64, err error) {
	f, err := os.Create(zipFile)
	if err != nil {
		return 0, err
	}
	defer func() { _ = f.Close() }()

	zipWriter := zip.NewWriter(f)
	defer func() { _ = zipWriter.Close() }()

	for _, file := range sourceFiles {
		if AddZipFile(file.Path, zipWriter) == nil {
			n++
		}
	}
	return n, nil
}

func AddZipFile(sourceFile string, zipWriter *zip.Writer) error {
	fileToZip, err := os.Open(sourceFile)
	if err != nil {
		log.Println("[osOpenError]", err.Error())
		return err
	}
	defer func() { _ = fileToZip.Close() }()

	// 获取文件信息
	fileInfo, err := fileToZip.Stat()
	if err != nil {
		log.Println("[fileToZipStatError]", err.Error())
		return err
	}
	// 创建一个新的 zip 文件头
	header, err := zip.FileInfoHeader(fileInfo)
	if err != nil {
		log.Println("[fileToZipHeaderError]", err.Error())
		return err
	}

	// 设置压缩方法为默认压缩
	header.Method = zip.Deflate

	// 添加文件头到压缩器中
	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		log.Println("[fileToZipCreateHeaderError]", err.Error())
		return err
	}
	// 将文件内容拷贝到压缩器中
	_, err = io.Copy(writer, fileToZip)
	if err != nil {
		log.Println("[fileToZipCopyError]", err.Error())
		return err
	}
	return nil
}
