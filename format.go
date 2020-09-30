package imgconv

import (
	"bytes"
	"image"
	"image/draw"
	"image/png"
	"io"
	"io/ioutil"
	"os"

	"github.com/disintegration/imaging"
	"github.com/sunshineplan/tiff"
)

// Format is an image file format.
// https://github.com/disintegration/imaging
type Format imaging.Format

// Image file formats.
const (
	JPEG Format = iota
	PNG
	GIF
	TIFF
	BMP
)

var formatExts = map[Format]string{
	JPEG: "jpg",
	PNG:  "png",
	GIF:  "gif",
	TIFF: "tif",
	BMP:  "bmp",
}

// FormatOption is format option
type FormatOption struct {
	Format       Format
	EncodeOption []EncodeOption
}

// EncodeOption sets an optional parameter for the Encode and Save functions.
// https://github.com/disintegration/imaging
type EncodeOption imaging.EncodeOption

// JPEGQuality returns an EncodeOption that sets the output JPEG quality.
// Quality ranges from 1 to 100 inclusive, higher is better.
func JPEGQuality(quality int) EncodeOption {
	return EncodeOption(imaging.JPEGQuality(quality))
}

// GIFNumColors returns an EncodeOption that sets the maximum number of colors
// used in the GIF-encoded image. It ranges from 1 to 256.  Default is 256.
func GIFNumColors(numColors int) EncodeOption {
	return EncodeOption(imaging.GIFNumColors(numColors))
}

// GIFQuantizer returns an EncodeOption that sets the quantizer that is used to produce
// a palette of the GIF-encoded image.
func GIFQuantizer(quantizer draw.Quantizer) EncodeOption {
	return EncodeOption(imaging.GIFQuantizer(quantizer))
}

// GIFDrawer returns an EncodeOption that sets the drawer that is used to convert
// the source image to the desired palette of the GIF-encoded image.
func GIFDrawer(drawer draw.Drawer) EncodeOption {
	return EncodeOption(imaging.GIFDrawer(drawer))
}

// PNGCompressionLevel returns an EncodeOption that sets the compression level
// of the PNG-encoded image. Default is png.DefaultCompression.
func PNGCompressionLevel(level png.CompressionLevel) EncodeOption {
	return EncodeOption(imaging.PNGCompressionLevel(level))
}

func setFormat(f string, options ...EncodeOption) (fo FormatOption, err error) {
	var format imaging.Format
	if format, err = imaging.FormatFromExtension(f); err != nil {
		return
	}
	fo.Format = Format(format)
	fo.EncodeOption = options
	return
}

func decode(r io.Reader, format Format) (image.Image, error) {
	if format == TIFF {
		// use forked tiff package because golang.org/x/image/tiff treat bad IFD tags order as invalid tiff
		return tiff.Decode(r)
	}
	return imaging.Decode(r)
}

// Decode reads an image from r.
func Decode(r io.Reader) (img image.Image, err error) {
	var b []byte
	b, err = ioutil.ReadAll(r)
	if err != nil {
		return
	}
	img, err = imaging.Decode(bytes.NewBuffer(b))
	if err != nil {
		// try forked tiff package
		img, err = tiff.Decode(bytes.NewBuffer(b))
	}
	return
}

// Open loads an image from file.
func Open(file string) (image.Image, error) {
	format, err := imaging.FormatFromFilename(file)
	if err != nil {
		format = -1
	}
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, err := decode(f, Format(format))
	if err != nil {
		return nil, err
	}
	return img, nil
}

// Write image according format option
func Write(base image.Image, w io.Writer, option FormatOption) error {
	return option.Write(base, w)
}

// Save saves image according format option
func Save(base image.Image, output string, option FormatOption) error {
	f, err := os.Create(output)
	if err != nil {
		return err
	}
	defer f.Close()
	return option.Write(base, f)
}

// Write image according format option
func (f *FormatOption) Write(base image.Image, w io.Writer) error {
	var opts []imaging.EncodeOption
	for _, i := range f.EncodeOption {
		opts = append(opts, imaging.EncodeOption(i))
	}
	return imaging.Encode(w, base, imaging.Format(f.Format), opts...)
}
