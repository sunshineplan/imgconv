package imgconv

import (
	"image"
	"io"
	"path/filepath"
)

const defaultOpacity = 128

var defaultFormat = &FormatOption{Format: JPEG}

// Options represents options that can be used to configure a image operation.
type Options struct {
	Watermark *WatermarkOption
	Resize    *ResizeOption
	Format    *FormatOption
	Gray      bool
}

// NewOptions creates a new option with default setting.
func NewOptions() *Options {
	return &Options{Format: defaultFormat}
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
func (opts *Options) SetFormat(f Format, options ...EncodeOption) *Options {
	opts.Format = &FormatOption{f, options}
	return opts
}

// SetGray sets the value for the Gray field.
func (opts *Options) SetGray(gray bool) *Options {
	opts.Gray = gray
	return opts
}

// Convert image according options opts.
func (opts *Options) Convert(w io.Writer, base image.Image) error {
	if opts.Gray {
		base = ToGray(base)
	}
	if opts.Resize != nil {
		base = opts.Resize.do(base)
	}
	if opts.Watermark != nil {
		base = opts.Watermark.do(base)
	}

	if opts.Format == nil {
		opts.Format = defaultFormat
	}

	return opts.Format.Encode(w, base)
}

// ConvertExt convert filename's ext according image format.
func (opts *Options) ConvertExt(filename string) string {
	return filename[0:len(filename)-len(filepath.Ext(filename))] + "." + formatExts[opts.Format.Format][0]
}
