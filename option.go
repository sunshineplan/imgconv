package img

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

var defaultFormat = format{format: imaging.JPEG, option: []imaging.EncodeOption{imaging.JPEGQuality(75)}}

// Option represents option that can be used to configure a image operation.
type Option struct {
	watermark *watermark
	resize    *resize
	format    format
}

// New return a default option.
func New() Option {
	return Option{format: defaultFormat}
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

// SetFormat sets the value for the Format field.
func (o *Option) SetFormat(f imaging.Format, option ...imaging.EncodeOption) *Option {
	o.format = format{format: f, option: option}
	return o
}

// Convert image by option
func (o *Option) Convert(src, dst string) error {
	output := o.format.path(dst)
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
	if o.resize != nil {
		img = o.resize.do(img)
	}
	if o.watermark != nil {
		img = o.watermark.do(img)
	}

	if err := os.MkdirAll(filepath.Dir(output), 0755); err != nil {
		return err
	}
	if err := o.format.save(img, output); err != nil {
		os.Remove(output)
		return err
	}

	return nil
}
