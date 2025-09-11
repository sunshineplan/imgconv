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
		mark, offset = w.randomWatermark(base.Bounds())
	} else {
		mark, offset = w.fixedWatermark(base.Bounds())
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

func (w *WatermarkOption) randomWatermark(base image.Rectangle) (image.Image, image.Point) {
	mark := w.Mark
	if mark.Bounds().Dx() >= base.Dx()/3 || mark.Bounds().Dy() >= base.Dy()/3 {
		opt := new(ResizeOption)
		if calcResizeXY(base, mark.Bounds()) {
			opt.Width = base.Dx() / 3
		} else {
			opt.Height = base.Dy() / 3
		}
		mark = Resize(mark, opt)
	}
	return rotate(mark, float64(randRange(-30, 30))+rand.Float64(), color.Transparent),
		image.Pt(
			randRange(base.Dx()/6, base.Dx()*5/6-mark.Bounds().Dx()),
			randRange(base.Dy()/6, base.Dy()*5/6-mark.Bounds().Dy()),
		)
}

func (w *WatermarkOption) fixedWatermark(base image.Rectangle) (image.Image, image.Point) {
	return w.Mark, image.Pt(
		(base.Bounds().Dx()/2)-(w.Mark.Bounds().Dx()/2)+w.Offset.X,
		(base.Bounds().Dy()/2)-(w.Mark.Bounds().Dy()/2)+w.Offset.Y,
	)
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
