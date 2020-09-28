package img

import (
	"image"
	"image/color"
	"image/draw"
	"math/rand"
	"time"

	"github.com/disintegration/imaging"
)

type watermark struct {
	mark    image.Image
	opacity uint8
	random  bool
	offset  image.Point
}

func (w *watermark) do(base image.Image) image.Image {
	output := image.NewRGBA(base.Bounds())
	draw.Draw(output, output.Bounds(), base, image.ZP, draw.Src)
	var offset image.Point
	var mark image.Image
	if w.random {
		rand.Seed(time.Now().UnixNano())
		if w.mark.Bounds().Dx() >= base.Bounds().Dx()/3 || w.mark.Bounds().Dy() >= base.Bounds().Dy()/3 {
			if calcResizeXY(base.Bounds(), w.mark.Bounds()) {
				mark = imaging.Resize(w.mark, base.Bounds().Dx()/3, 0, imaging.Lanczos)
			} else {
				mark = imaging.Resize(w.mark, 0, base.Bounds().Dy()/3, imaging.Lanczos)
			}
		} else {
			mark = w.mark
		}
		mark = imaging.Rotate(mark, float64(randRange(-30, 30))+rand.Float64(), color.Transparent)
		offset = randOffset(base.Bounds(), mark.Bounds())
	} else {
		mark = w.mark
		offset = calcOffset(base.Bounds(), mark.Bounds(), w.offset)
	}
	draw.DrawMask(output, mark.Bounds().Add(offset), mark, image.ZP, image.NewUniform(color.Alpha{w.opacity}), image.ZP, draw.Over)
	return output
}

func randRange(min, max int) int {
	return rand.Intn(max-min+1) + min
}

func randOffset(base, mark image.Rectangle) image.Point {
	return image.Pt(
		randRange(base.Bounds().Dx()/6, base.Bounds().Dx()*5/6-mark.Bounds().Dx()),
		randRange(base.Bounds().Dy()/6, base.Bounds().Dy()*5/6-mark.Bounds().Dy()))
}

func calcOffset(base, mark image.Rectangle, p image.Point) image.Point {
	return image.Pt(
		(base.Size().X/2)-(mark.Size().X/2)+p.X,
		(base.Size().Y/2)-(mark.Size().Y/2)+p.Y)
}

func calcResizeXY(base, mark image.Rectangle) bool {
	if base.Dx()*mark.Dy()/mark.Dx() < base.Dy() {
		return true
	}
	return false
}
