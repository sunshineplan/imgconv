package imgconv

import (
	"encoding"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"strings"

	"github.com/sunshineplan/pdf"
	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"
	_ "golang.org/x/image/webp" // decode webp format
)

var (
	_ encoding.TextUnmarshaler = new(Format)
	_ encoding.TextMarshaler   = Format(0)
)

// Format is an image file format.
type Format int

// Image file formats.
const (
	JPEG Format = iota
	PNG
	GIF
	TIFF
	BMP
	PDF
)

var formatExts = [][]string{
	{"jpg", "jpeg"},
	{"png"},
	{"gif"},
	{"tif", "tiff"},
	{"bmp"},
	{"pdf"},
}

func (f Format) String() (format string) {
	defer func() {
		if err := recover(); err != nil {
			format = "unknown"
		}
	}()
	return formatExts[f][0]
}

// FormatFromExtension parses image format from filename extension:
// "jpg" (or "jpeg"), "png", "gif", "tif" (or "tiff"), "bmp" and "pdf" are supported.
func FormatFromExtension(ext string) (Format, error) {
	ext = strings.ToLower(ext)
	for index, exts := range formatExts {
		for _, i := range exts {
			if ext == i {
				return Format(index), nil
			}
		}
	}

	return -1, image.ErrFormat
}

func (f *Format) UnmarshalText(text []byte) error {
	format, err := FormatFromExtension(string(text))
	if err != nil {
		return err
	}
	*f = format
	return nil
}

func (f Format) MarshalText() ([]byte, error) {
	return []byte(f.String()), nil
}

var (
	_ encoding.TextUnmarshaler = new(TIFFCompression)
	_ encoding.TextMarshaler   = TIFFCompression(0)
)

// TIFFCompression describes the type of compression used in Options.
type TIFFCompression int

// Constants for supported TIFF compression types.
const (
	TIFFUncompressed TIFFCompression = iota
	TIFFDeflate
)

var tiffCompression = []string{
	"none",
	"deflate",
}

func (c TIFFCompression) value() tiff.CompressionType {
	switch c {
	case TIFFDeflate:
		return tiff.Deflate
	}
	return tiff.Uncompressed
}

func (c *TIFFCompression) UnmarshalText(text []byte) error {
	t := strings.ToLower(string(text))
	for index, tt := range tiffCompression {
		if t == tt {
			*c = TIFFCompression(index)
			return nil
		}
	}
	return fmt.Errorf("tiff: unsupported compression: %s", t)
}

func (c TIFFCompression) MarshalText() (b []byte, err error) {
	defer func() {
		if err := recover(); err != nil {
			b = []byte("unknown")
		}
	}()
	ct := tiffCompression[c]
	return []byte(ct), nil
}

// FormatOption is format option
type FormatOption struct {
	Format       Format
	EncodeOption []EncodeOption
}

type encodeConfig struct {
	Quality             int
	gifNumColors        int
	gifQuantizer        draw.Quantizer
	gifDrawer           draw.Drawer
	pngCompressionLevel png.CompressionLevel
	tiffCompressionType TIFFCompression
	background          color.Color
}

var defaultEncodeConfig = encodeConfig{
	Quality:             75,
	gifNumColors:        256,
	gifQuantizer:        nil,
	gifDrawer:           nil,
	pngCompressionLevel: png.DefaultCompression,
	tiffCompressionType: TIFFDeflate,
}

// EncodeOption sets an optional parameter for the Encode and Save functions.
// https://github.com/disintegration/imaging
type EncodeOption func(*encodeConfig)

// Quality returns an EncodeOption that sets the output JPEG or PDF quality.
// Quality ranges from 1 to 100 inclusive, higher is better.
func Quality(quality int) EncodeOption {
	return func(c *encodeConfig) {
		c.Quality = quality
	}
}

// GIFNumColors returns an EncodeOption that sets the maximum number of colors
// used in the GIF-encoded image. It ranges from 1 to 256.  Default is 256.
func GIFNumColors(numColors int) EncodeOption {
	return func(c *encodeConfig) {
		c.gifNumColors = numColors
	}
}

// GIFQuantizer returns an EncodeOption that sets the quantizer that is used to produce
// a palette of the GIF-encoded image.
func GIFQuantizer(quantizer draw.Quantizer) EncodeOption {
	return func(c *encodeConfig) {
		c.gifQuantizer = quantizer
	}
}

// GIFDrawer returns an EncodeOption that sets the drawer that is used to convert
// the source image to the desired palette of the GIF-encoded image.
func GIFDrawer(drawer draw.Drawer) EncodeOption {
	return func(c *encodeConfig) {
		c.gifDrawer = drawer
	}
}

// PNGCompressionLevel returns an EncodeOption that sets the compression level
// of the PNG-encoded image. Default is png.DefaultCompression.
func PNGCompressionLevel(level png.CompressionLevel) EncodeOption {
	return func(c *encodeConfig) {
		c.pngCompressionLevel = level
	}
}

// TIFFCompressionType returns an EncodeOption that sets the compression type
// of the TIFF-encoded image. Default is tiff.Deflate.
func TIFFCompressionType(compressionType TIFFCompression) EncodeOption {
	return func(c *encodeConfig) {
		c.tiffCompressionType = compressionType
	}
}

// BackgroundColor returns an EncodeOption that sets the background color.
func BackgroundColor(color color.Color) EncodeOption {
	return func(c *encodeConfig) {
		c.background = color
	}
}

// Encode writes the image img to w in the specified format (JPEG, PNG, GIF, TIFF, BMP or PDF).
func (f *FormatOption) Encode(w io.Writer, img image.Image) error {
	cfg := defaultEncodeConfig
	for _, option := range f.EncodeOption {
		option(&cfg)
	}

	if cfg.background != nil {
		i := image.NewNRGBA(img.Bounds())
		draw.Draw(i, i.Bounds(), &image.Uniform{cfg.background}, img.Bounds().Min, draw.Src)
		draw.Draw(i, i.Bounds(), img, img.Bounds().Min, draw.Over)
		img = i
	}

	switch f.Format {
	case JPEG:
		if nrgba, ok := img.(*image.NRGBA); ok && nrgba.Opaque() {
			rgba := &image.RGBA{
				Pix:    nrgba.Pix,
				Stride: nrgba.Stride,
				Rect:   nrgba.Rect,
			}
			return jpeg.Encode(w, rgba, &jpeg.Options{Quality: cfg.Quality})
		}
		return jpeg.Encode(w, img, &jpeg.Options{Quality: cfg.Quality})

	case PNG:
		encoder := png.Encoder{CompressionLevel: cfg.pngCompressionLevel}
		return encoder.Encode(w, img)

	case GIF:
		return gif.Encode(w, img, &gif.Options{
			NumColors: cfg.gifNumColors,
			Quantizer: cfg.gifQuantizer,
			Drawer:    cfg.gifDrawer,
		})

	case TIFF:
		return tiff.Encode(w, img, &tiff.Options{Compression: cfg.tiffCompressionType.value(), Predictor: true})

	case BMP:
		return bmp.Encode(w, img)

	case PDF:
		return pdf.Encode(w, []image.Image{img}, &pdf.Options{Quality: cfg.Quality})
	}

	return image.ErrFormat
}
