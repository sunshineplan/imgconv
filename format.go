package img

import (
	"image"
	"path/filepath"

	"github.com/disintegration/imaging"
)

var formatExts = map[imaging.Format]string{
	imaging.JPEG: ".jpg",
	imaging.PNG:  ".png",
	imaging.GIF:  ".gif",
	imaging.TIFF: ".tif",
	imaging.BMP:  ".bmp",
}

type format struct {
	format imaging.Format
	option []imaging.EncodeOption
}

func (f *format) save(base image.Image, output string) error {
	return imaging.Save(base, output, f.option...)
}

func (f *format) path(dst string) string {
	return dst[0:len(dst)-len(filepath.Ext(dst))] + formatExts[f.format]
}
