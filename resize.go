package imgconv

import (
	"image"

	"github.com/disintegration/imaging"
)

// ResizeOption is resize option
type ResizeOption struct {
	Width   int
	Height  int
	Percent float64
}

// Resize resizes image
func Resize(base image.Image, option *ResizeOption) image.Image {
	return option.do(base)
}

func (r *ResizeOption) do(base image.Image) image.Image {
	if r.Width == 0 && r.Height == 0 {
		return imaging.Resize(base, int(float64(base.Bounds().Dx())*r.Percent/100), 0, imaging.Lanczos)
	}

	return imaging.Resize(base, r.Width, r.Height, imaging.Lanczos)
}
