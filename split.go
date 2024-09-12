package imgconv

import (
	"errors"
	"image"
)

type SplitMode int

const (
	SplitHorizontalMode SplitMode = iota
	SplitVerticalMode
)

func split(base image.Rectangle, n int, mode SplitMode) (rects []image.Rectangle) {
	var width, height int
	if mode == SplitHorizontalMode {
		width = base.Bounds().Dx() / n
		height = base.Bounds().Dy()
	} else {
		width = base.Bounds().Dx()
		height = base.Bounds().Dy() / n
	}
	if width == 0 || height == 0 {
		return
	}
	for i := range n {
		var r image.Rectangle
		if mode == SplitHorizontalMode {
			r = image.Rect(
				base.Bounds().Min.X+width*i, base.Bounds().Min.Y,
				base.Bounds().Min.X+width*(i+1), base.Bounds().Min.Y+height,
			)
		} else {
			r = image.Rect(
				base.Bounds().Min.X, base.Bounds().Min.Y+height*i,
				base.Bounds().Min.X+width, base.Bounds().Min.Y+height*(i+1),
			)
		}
		rects = append(rects, r)
	}
	return
}

func Split(base image.Image, n int, mode SplitMode) (imgs []image.Image, err error) {
	if n < 1 {
		return nil, errors.New("")
	}
	if img, ok := base.(interface {
		SubImage(image.Rectangle) image.Image
	}); ok {
		rects := split(base.Bounds(), n, mode)
		if len(rects) == 0 {
			return nil, errors.New("")
		}
		for _, rect := range rects {
			imgs = append(imgs, img.SubImage(rect))
		}
	} else {
		return nil, errors.New("")
	}
	return
}

func SplitHorizontal(base image.Image, n int) ([]image.Image, error) {
	return Split(base, n, SplitHorizontalMode)
}

func SplitVertical(base image.Image, n int) ([]image.Image, error) {
	return Split(base, n, SplitVerticalMode)
}
