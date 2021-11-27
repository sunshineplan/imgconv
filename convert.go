package imgconv

import (
	"bytes"
	"image"
	"io"
	"os"

	"github.com/sunshineplan/tiff"
)

// Decode reads an image from r.
func Decode(r io.Reader) (image.Image, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	img, format, err := image.Decode(bytes.NewBuffer(b))
	if format == "tiff" && err != nil {
		return tiff.Decode(bytes.NewBuffer(b))
	}

	return img, err
}

// DecodeConfig decodes the color model and dimensions of an image that has been encoded in a
// registered format. The string returned is the format name used during format registration.
func DecodeConfig(r io.Reader) (image.Config, string, error) {
	return image.DecodeConfig(r)
}

// Open loads an image from file.
func Open(file string) (image.Image, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return Decode(f)
}

// Write image according format option
func Write(w io.Writer, base image.Image, option FormatOption) error {
	return option.Encode(w, base)
}

// Save saves image according format option
func Save(output string, base image.Image, option FormatOption) error {
	f, err := os.Create(output)
	if err != nil {
		return err
	}
	defer f.Close()

	return option.Encode(f, base)
}
