package handlers

import (
	"github.com/davidbyttow/govips/v2/vips"
	"github.com/lixiang4u/imago/models"
)

func ImageFilter(img *vips.ImageRef, config models.ImageConfig, appConfig models.AppConfig) *vips.ImageRef {

	return img
}

func ExportImage(img *vips.ImageRef, toType vips.ImageType, exportParams models.ExportConfig) (buf []byte, meta *vips.ImageMetadata, err error) {
	switch toType {
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
		buf, meta, err = img.ExportWebp(&vips.WebpExportParams{
			StripMetadata:   exportParams.StripMetadata,
			Lossless:        exportParams.Lossless,
			Quality:         exportParams.Quality,
			ReductionEffort: exportParams.ReductionEffort,
		})
	}
	return
}
