package imgconv

import (
	"image"
	"image/color"
	"image/draw"
	"math/rand"
	"time"

	"github.com/disintegration/imaging"
)

// WatermarkOption is watermark option
type WatermarkOption struct {
	mark    string
	Mark    image.Image
	Opacity uint8
	Random  bool
	Offset  image.Point
}

// Watermark add watermark to image
func Watermark(base image.Image, option WatermarkOption) image.Image {
	return option.do(base)
}

func (w *WatermarkOption) do(base image.Image) image.Image {
	output := image.NewRGBA(base.Bounds())
	draw.Draw(output, output.Bounds(), base, image.ZP, draw.Src)
	var offset image.Point
	var mark image.Image
	if w.Random {
		rand.Seed(time.Now().UnixNano())
		if w.Mark.Bounds().Dx() >= base.Bounds().Dx()/3 || w.Mark.Bounds().Dy() >= base.Bounds().Dy()/3 {
			if calcResizeXY(base.Bounds(), w.Mark.Bounds()) {
				mark = imaging.Resize(w.Mark, base.Bounds().Dx()/3, 0, imaging.Lanczos)
			} else {
				mark = imaging.Resize(w.Mark, 0, base.Bounds().Dy()/3, imaging.Lanczos)
			}
		} else {
			mark = w.Mark
		}
		mark = imaging.Rotate(mark, float64(randRange(-30, 30))+rand.Float64(), color.Transparent)
		offset = image.Pt(
			randRange(base.Bounds().Dx()/6, base.Bounds().Dx()*5/6-mark.Bounds().Dx()),
			randRange(base.Bounds().Dy()/6, base.Bounds().Dy()*5/6-mark.Bounds().Dy()))
	} else {
		mark = w.Mark
		offset = image.Pt(
			(base.Bounds().Dx()/2)-(mark.Bounds().Dx()/2)+w.Offset.X,
			(base.Bounds().Dy()/2)-(mark.Bounds().Dy()/2)+w.Offset.Y)
	}
	draw.DrawMask(output, mark.Bounds().Add(offset), mark, image.ZP, image.NewUniform(color.Alpha{w.Opacity}), image.ZP, draw.Over)
	return output
}

func randRange(min, max int) int {
	return rand.Intn(max-min+1) + min
}

func calcResizeXY(base, mark image.Rectangle) bool {
	if base.Dx()*mark.Dy()/mark.Dx() < base.Dy() {
		return true
	}
	return false
}
