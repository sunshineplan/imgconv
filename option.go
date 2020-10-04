package imgconv

import (
	"image"
	"io"
	"path/filepath"
	"reflect"
)

const defaultOpacity = 128

var defaultFormat = FormatOption{Format: JPEG, EncodeOption: []EncodeOption{JPEGQuality(75)}}

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
func (o *Options) SetWatermark(mark image.Image, opacity uint) *Options {
	o.Watermark = &WatermarkOption{Mark: mark}
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
func (o *Options) SetFormat(f string, options ...EncodeOption) (err error) {
	o.Format, err = setFormat(f, options...)
	return
}

// Convert image by options
func (o *Options) Convert(w io.Writer, base image.Image) error {
	if o.Resize != nil {
		base = o.Resize.do(base)
	}
	if o.Watermark != nil {
		base = o.Watermark.do(base)
	}

	if reflect.DeepEqual(o.Format, FormatOption{}) {
		o.Format = defaultFormat
	}
	return o.Format.Write(w, base)
}

// ConvertExt convert filename's ext according image format
func (o *Options) ConvertExt(filename string) string {
	return filename[0:len(filename)-len(filepath.Ext(filename))] + "." + formatExts[o.Format.Format]
}
