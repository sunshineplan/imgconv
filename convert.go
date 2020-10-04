package imgconv

import (
	"image"
	"io"
	"os"
)

// Decode reads an image from r.
func Decode(r io.Reader) (image.Image, error) {
	img, _, err := image.Decode(r)
	return img, err
}

// DecodeConfig returns the color model and dimensions of a image without
// decoding the entire image.
func DecodeConfig(r io.Reader) (image.Config, error) {
	cfg, _, err := image.DecodeConfig(r)
	return cfg, err
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
