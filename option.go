package imgconv

import (
	"image"
	"io"
	"path/filepath"
	"reflect"
)

const defaultOpacity = 128

var defaultFormat = FormatOption{Format: JPEG}

// Options represents options that can be used to configure a image operation.
type Options struct {
	Watermark *WatermarkOption
	Resize    *ResizeOption
	Format    FormatOption
}

// NewOptions creates a new option with default setting.
func NewOptions() Options {
	return Options{Format: defaultFormat}
}

// SetWatermark sets the value for the Watermark field.
func (opts *Options) SetWatermark(mark image.Image, opacity uint) *Options {
	opts.Watermark = &WatermarkOption{Mark: mark}
	if opacity == 0 {
		opts.Watermark.Opacity = defaultOpacity
	} else {
		opts.Watermark.Opacity = uint8(opacity)
	}

	return opts
}

// SetResize sets the value for the Resize field.
func (opts *Options) SetResize(width, height int, percent float64) *Options {
	opts.Resize = &ResizeOption{Width: width, Height: height, Percent: percent}
	return opts
}

// SetFormat sets the value for the Format field.
func (opts *Options) SetFormat(f string, options ...EncodeOption) (err error) {
	opts.Format, err = setFormat(f, options...)
	return
}

// Convert image according options opts.
func (opts *Options) Convert(w io.Writer, base image.Image) error {
	if opts.Resize != nil {
		base = opts.Resize.do(base)
	}
	if opts.Watermark != nil {
		base = opts.Watermark.do(base)
	}

	if reflect.DeepEqual(opts.Format, FormatOption{}) {
		opts.Format = defaultFormat
	}

	return opts.Format.Encode(w, base)
}

// ConvertExt convert filename's ext according image format.
func (opts *Options) ConvertExt(filename string) string {
	return filename[0:len(filename)-len(filepath.Ext(filename))] + "." + formatExts[opts.Format.Format]
}
