package imgconv

import (
	"image"
	"image/color"
	"image/draw"
	"math/rand/v2"
)

// WatermarkOption is watermark option
type WatermarkOption struct {
	Mark    image.Image
	Opacity uint8
	Random  bool
	Offset  image.Point
}

// Watermark add watermark to image
func Watermark(base image.Image, option *WatermarkOption) image.Image {
	return option.do(base)
}

// SetRandom sets the option for the Watermark position random or not.
func (w *WatermarkOption) SetRandom(random bool) *WatermarkOption {
	w.Random = random
	return w
}

// SetOffset sets the option for the Watermark offset base center when adding fixed watermark.
func (w *WatermarkOption) SetOffset(offset image.Point) *WatermarkOption {
	w.Offset = offset
	return w
}

func (w *WatermarkOption) do(base image.Image) image.Image {
	img := image.NewNRGBA(base.Bounds())
	draw.Draw(img, img.Bounds(), base, image.Point{}, draw.Src)
	var offset image.Point
	var mark image.Image
	if w.Random {
		if w.Mark.Bounds().Dx() >= base.Bounds().Dx()/3 || w.Mark.Bounds().Dy() >= base.Bounds().Dy()/3 {
			if calcResizeXY(base.Bounds(), w.Mark.Bounds()) {
				mark = Resize(w.Mark, &ResizeOption{Width: base.Bounds().Dx() / 3})
			} else {
				mark = Resize(w.Mark, &ResizeOption{Height: base.Bounds().Dy() / 3})
			}
		} else {
			mark = w.Mark
		}
		mark = rotate(mark, float64(randRange(-30, 30))+rand.Float64(), color.Transparent)
		offset = image.Pt(
			randRange(base.Bounds().Dx()/6, base.Bounds().Dx()*5/6-mark.Bounds().Dx()),
			randRange(base.Bounds().Dy()/6, base.Bounds().Dy()*5/6-mark.Bounds().Dy()))
	} else {
		mark = w.Mark
		offset = image.Pt(
			(base.Bounds().Dx()/2)-(mark.Bounds().Dx()/2)+w.Offset.X,
			(base.Bounds().Dy()/2)-(mark.Bounds().Dy()/2)+w.Offset.Y)
	}

	draw.DrawMask(
		img,
		mark.Bounds().Add(offset),
		mark,
		image.Point{},
		image.NewUniform(color.Alpha{w.Opacity}),
		image.Point{},
		draw.Over,
	)

	return img
}

func randRange(min, max int) int {
	if max < min {
		min, max = max, min
	}
	return rand.N(max-min+1) + min
}

func calcResizeXY(base, mark image.Rectangle) bool {
	return base.Dx()*mark.Dy()/mark.Dx() < base.Dy()
}
