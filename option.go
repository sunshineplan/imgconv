package imgconv

import (
	"image"
	"io"
	"reflect"

	"github.com/disintegration/imaging"
)

const defaultOpacity = 128

var defaultFormat = formatOption{format: imaging.JPEG, encodeOption: []imaging.EncodeOption{imaging.JPEGQuality(75)}}

// Options represents options that can be used to configure a image operation.
type Options struct {
	Watermark *WatermarkOption
	Resize    *ResizeOption
	format    formatOption
}

// New return a default option.
func New() Options {
	return Options{format: defaultFormat}
}

// SetWatermark sets the value for the Watermark field.
func (o *Options) SetWatermark(mark string, opacity uint, random bool, offset image.Point) (*Options, error) {
	img, err := imaging.Open(mark)
	if err != nil {
		return nil, err
	}
	o.Watermark = &WatermarkOption{mark: mark, Mark: img, Random: random}
	if !random {
		o.Watermark.Offset = offset
	}
	if opacity == 0 {
		o.Watermark.Opacity = defaultOpacity
	} else {
		o.Watermark.Opacity = uint8(opacity)
	}
	return o, nil
}

// SetResize sets the value for the Resize field.
func (o *Options) SetResize(width, height int, percent float64) *Options {
	o.Resize = &ResizeOption{Width: width, Height: height, Percent: percent}
	return o
}

// SetFormat sets the value for the Format field.
func (o *Options) SetFormat(f string, options ...imaging.EncodeOption) error {
	var format imaging.Format
	var err error
	if format, err = imaging.FormatFromExtension(f); err != nil {
		return err
	}
	o.format = formatOption{format: format, encodeOption: options}
	return nil
}

// Convert image by option
func (o *Options) Convert(base image.Image, w io.Writer) error {
	//output := o.format.path(dst)
	//if _, err := os.Stat(output); !os.IsNotExist(err) {
	//	return os.ErrExist
	//}

	//var img, err := Open(src)
	//if err != nil {
	//	return err
	//}
	if o.Resize != nil {
		base = o.Resize.do(base)
	}
	if o.Watermark != nil {
		base = o.Watermark.do(base)
	}

	if reflect.DeepEqual(o.format, formatOption{}) {
		o.format = defaultFormat
	}
	//if err := os.MkdirAll(filepath.Dir(output), 0755); err != nil {
	//	return err
	//}
	return o.format.encode(base, w)
	//os.Remove(output)
}
