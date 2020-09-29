package imgconv

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

// FormatOption is format option
type FormatOption struct {
	Format       imaging.Format
	EncodeOption []imaging.EncodeOption
}

// Export saves image according FormatOption
func Export(base image.Image, output string, option FormatOption) error {
	return option.save(base, output)
}

func (f *FormatOption) save(base image.Image, output string) error {
	return imaging.Save(base, output, f.EncodeOption...)
}

func (f *FormatOption) path(dst string) string {
	return dst[0:len(dst)-len(filepath.Ext(dst))] + formatExts[f.Format]
}
