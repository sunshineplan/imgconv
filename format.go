package imgconv

import (
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/sunshineplan/pdf"
	"github.com/sunshineplan/tiff" // decode tiff format, not check IFD tags order
	"golang.org/x/image/bmp"
	_ "golang.org/x/image/webp" // decode webp format
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

func (f Format) String() string {
	return formatExts[f][0]
}

// FormatFromExtension parses image format from filename extension:
// "jpg" (or "jpeg"), "png", "gif", "tif" (or "tiff"), "bmp" and "pdf" are supported.
func FormatFromExtension(ext string) (Format, error) {
	ext = strings.ToLower(ext)
	for _, exts := range formatExts {
		for index, i := range exts {
			if ext == i {
				return Format(index), nil
			}
		}
	}

	return -1, image.ErrFormat
}

// TIFFCompression describes the type of compression used in Options.
type TIFFCompression int

// Constants for supported TIFF compression types.
const (
	TIFFUncompressed TIFFCompression = iota
	TIFFDeflate
	TIFFLZW
	TIFFCCITTGroup3
	TIFFCCITTGroup4
	TIFFJPEG
)

func (c TIFFCompression) value() tiff.CompressionType {
	switch c {
	case TIFFLZW:
		return tiff.LZW
	case TIFFDeflate:
		return tiff.Deflate
	case TIFFCCITTGroup3:
		return tiff.CCITTGroup3
	case TIFFCCITTGroup4:
		return tiff.CCITTGroup4
	case TIFFJPEG:
		return tiff.JPEG
	}
	return tiff.Uncompressed
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
	tiffCompressionType: TIFFLZW,
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
		i := imaging.Clone(img)
		draw.Draw(i, i.Bounds(), &image.Uniform{cfg.background}, image.Point{}, draw.Src)
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
