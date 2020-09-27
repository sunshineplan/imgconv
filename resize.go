package img

import (
	"image"

	"github.com/disintegration/imaging"
)

type resize struct {
	width   int
	height  int
	percent float64
}

func (r resize) do(base image.Image) image.Image {
	if r.width == 0 && r.height == 0 {
		return imaging.Resize(base, int(float64(base.Bounds().Dx())*r.percent/100), 0, imaging.Lanczos)
	}
	return imaging.Resize(base, r.width, r.height, imaging.Lanczos)
}
