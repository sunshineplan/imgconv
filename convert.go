package imgconv

import (
	"image"
	"io"
	"os"
)

type decodeConfig struct {
	autoOrientation bool
}

var defaultDecodeConfig = decodeConfig{
	autoOrientation: true,
}

// DecodeOption sets an optional parameter for the Decode and Open functions.
type DecodeOption func(*decodeConfig)

// AutoOrientation returns a DecodeOption that sets the auto-orientation mode.
// If auto-orientation is enabled, the image will be transformed after decoding
// according to the EXIF orientation tag (if present). By default it's enabled.
func AutoOrientation(enabled bool) DecodeOption {
	return func(c *decodeConfig) {
		c.autoOrientation = enabled
	}
}

// Decode reads an image from r.
// If want to use custom image format packages which were registered in image package, please
// make sure these custom packages imported before importing imgconv package.
func Decode(r io.Reader, opts ...DecodeOption) (image.Image, error) {
	cfg := defaultDecodeConfig
	for _, option := range opts {
		option(&cfg)
	}

	return decode(r, autoOrientation(cfg.autoOrientation))
}

// DecodeConfig decodes the color model and dimensions of an image that has been encoded in a
// registered format. The string returned is the format name used during format registration.
func DecodeConfig(r io.Reader) (image.Config, string, error) {
	return image.DecodeConfig(r)
}

// Open loads an image from file.
func Open(file string, opts ...DecodeOption) (image.Image, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return Decode(f, opts...)
}

// Write image according format option
func Write(w io.Writer, base image.Image, option *FormatOption) error {
	return option.Encode(w, base)
}

// Save saves image according format option
func Save(output string, base image.Image, option *FormatOption) error {
	f, err := os.Create(output)
	if err != nil {
		return err
	}
	defer f.Close()

	return option.Encode(f, base)
}
