package img

import (
	"image"
	"image/jpeg"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/sunshineplan/tiff"
)

const (
	defaultOpacity = 128
	defaultQuality = 75
)

// Option represents option that can be used to configure a image operation.
type Option struct {
	watermark *watermark
	resize    *resize
	Quality   int
}

// New return a default option.
func New() Option {
	return Option{Quality: defaultQuality}
}

// SetWatermark sets the value for the Watermark field.
func (o *Option) SetWatermark(mark string, opacity uint, random bool, offset image.Point) *Option {
	img, err := imaging.Open(mark)
	if err != nil {
		log.Fatal(err)
	}
	o.watermark = &watermark{mark: img, random: random}
	if !random {
		o.watermark.offset = offset
	}
	if opacity == 0 {
		o.watermark.opacity = defaultOpacity
	} else {
		o.watermark.opacity = uint8(opacity)
	}
	return o
}

// SetResize sets the value for the Resize field.
func (o *Option) SetResize(width, height int, percent float64) *Option {
	o.resize = &resize{width: width, height: height, percent: percent}
	return o
}

// Test if option is runnable.
func (o *Option) Test() bool {
	if o.watermark == nil && o.resize == nil {
		return false
	}
	return true
}

// Convert image by option
func (o *Option) Convert(src, dst string) error {
	if _, err := os.Stat(dst); !os.IsNotExist(err) {
		return os.ErrExist
	}

	var img image.Image
	var err error
	if ext := strings.ToLower(filepath.Ext(src)); ext == ".tif" || ext == ".tiff" {
		f, err := os.Open(src)
		if err != nil {
			return err
		}
		defer f.Close()
		img, err = tiff.Decode(f)
	} else {
		img, err = imaging.Open(src)
	}
	if err != nil {
		return err
	}
	if o.resize != nil {
		img = o.resize.do(img)
	}
	if o.watermark != nil {
		img = o.watermark.do(img)
	}

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	err = jpeg.Encode(f, img, &jpeg.Options{Quality: o.Quality})
	f.Close()
	if err != nil {
		os.Remove(dst)
	}

	return err
}
