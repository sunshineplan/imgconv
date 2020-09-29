package imgconv

import (
	"fmt"
	"image"
	"os"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
	"golang.org/x/image/tiff"
)

var formatExts = map[imaging.Format]string{
	imaging.JPEG: ".jpg",
	imaging.PNG:  ".png",
	imaging.GIF:  ".gif",
	imaging.TIFF: ".tif",
	imaging.BMP:  ".bmp",
}

var tiffCompressionType = map[string]tiff.CompressionType{
	"uncompressed": tiff.Uncompressed,
	"deflate":      tiff.Deflate,
	"lzw":          tiff.LZW,
	"ccitt3":       tiff.CCITTGroup3,
	"ccitt4":       tiff.CCITTGroup4,
}

type formatOption struct {
	format       imaging.Format
	encodeOption []interface{}
}

// Export saves image according FormatOption
func Export(base image.Image, output string, option formatOption) error {
	return option.save(base, output)
}

func (f *formatOption) save(base image.Image, output string) error {
	if f.format == imaging.TIFF {
		file, err := os.Create(output)
		if err != nil {
			return err
		}
		defer file.Close()
		opt := f.encodeOption[0].(tiff.Options)
		return tiff.Encode(file, base, &opt)
	}
	var opts []imaging.EncodeOption
	for _, i := range f.encodeOption {
		opts = append(opts, i.(imaging.EncodeOption))
	}
	return imaging.Save(base, output, opts...)
}

func (f *formatOption) path(dst string) string {
	return dst[0:len(dst)-len(filepath.Ext(dst))] + formatExts[f.format]
}

// ParseTIFFCompressionType parse tiff compression type
func ParseTIFFCompressionType(t string) (tiff.Options, error) {
	if compression, ok := tiffCompressionType[strings.ToLower(t)]; ok {
		return tiff.Options{Compression: compression}, nil
	}
	return tiff.Options{}, fmt.Errorf("unsupported tiff compression type: %s", t)
}
