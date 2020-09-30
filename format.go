package imgconv

import (
	"image"
	"io"
	"os"
	"path/filepath"

	"github.com/disintegration/imaging"
	tif "github.com/sunshineplan/tiff"
)

var formatExts = map[imaging.Format]string{
	imaging.JPEG: ".jpg",
	imaging.PNG:  ".png",
	imaging.GIF:  ".gif",
	imaging.TIFF: ".tif",
	imaging.BMP:  ".bmp",
}

type formatOption struct {
	format       imaging.Format
	encodeOption []imaging.EncodeOption
}

func decode(r io.Reader, format imaging.Format) (image.Image, error) {
	if format == imaging.TIFF {
		return tif.Decode(r)
	}
	return imaging.Decode(r)
}

// Open image according given format
func Open(src string) (image.Image, error) {
	format, err := imaging.FormatFromFilename(src)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(src)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, err := decode(f, format)
	if err != nil {
		return nil, err
	}
	return img, nil
}

// Encode image according format option
func Encode(base image.Image, w io.Writer, option formatOption) error {
	return option.encode(base, w)
}

// Export saves image according format option
func Export(base image.Image, output string, option formatOption) error {
	return option.save(base, output)
}

func (f *formatOption) encode(base image.Image, w io.Writer) error {
	return imaging.Encode(w, base, f.format, f.encodeOption...)
}

func (f *formatOption) save(base image.Image, output string) error {
	file, err := os.Create(output)
	if err != nil {
		return err
	}
	defer file.Close()
	return f.encode(base, file)
}

func (f *formatOption) path(dst string) string {
	return dst[0:len(dst)-len(filepath.Ext(dst))] + formatExts[f.format]
}
