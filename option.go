package imgconv

import (
	"image"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/sunshineplan/tiff"
)

const defaultOpacity = 128

var defaultFormat = FormatOption{Format: imaging.JPEG, EncodeOption: []imaging.EncodeOption{imaging.JPEGQuality(75)}}

// Options represents options that can be used to configure a image operation.
type Options struct {
	Watermark *WatermarkOption
	Resize    *ResizeOption
	Format    FormatOption
}

// New return a default option.
func New() Options {
	return Options{Format: defaultFormat}
}

// SetWatermark sets the value for the Watermark field.
func (o *Options) SetWatermark(mark string, opacity uint, random bool, offset image.Point) *Options {
	img, err := imaging.Open(mark)
	if err != nil {
		log.Fatal(err)
	}
	o.Watermark = &WatermarkOption{Mark: img, Random: random}
	if !random {
		o.Watermark.Offset = offset
	}
	if opacity == 0 {
		o.Watermark.Opacity = defaultOpacity
	} else {
		o.Watermark.Opacity = uint8(opacity)
	}
	return o
}

// SetResize sets the value for the Resize field.
func (o *Options) SetResize(width, height int, percent float64) *Options {
	o.Resize = &ResizeOption{Width: width, Height: height, Percent: percent}
	return o
}

// SetFormat sets the value for the Format field.
func (o *Options) SetFormat(f imaging.Format, option ...imaging.EncodeOption) *Options {
	o.Format = FormatOption{Format: f, EncodeOption: option}
	return o
}

// Convert image by option
func (o *Options) Convert(src, dst string) error {
	output := o.Format.path(dst)
	if _, err := os.Stat(output); !os.IsNotExist(err) {
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
	if o.Resize != nil {
		img = o.Resize.do(img)
	}
	if o.Watermark != nil {
		img = o.Watermark.do(img)
	}

	if err := os.MkdirAll(filepath.Dir(output), 0755); err != nil {
		return err
	}
	if err := o.Format.save(img, output); err != nil {
		os.Remove(output)
		return err
	}

	return nil
}
