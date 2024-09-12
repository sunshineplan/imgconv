package imgconv

import (
	"errors"
	"image"
)

// SplitMode defines the mode in which the image will be split
type SplitMode int

const (
	// SplitHorizontalMode splits the image horizontally
	SplitHorizontalMode SplitMode = iota
	// SplitVerticalMode splits the image vertically
	SplitVerticalMode
)

func split(base image.Rectangle, n int, mode SplitMode) (rects []image.Rectangle) {
	var width, height int
	if mode == SplitHorizontalMode {
		width = base.Dx() / n
		height = base.Dy()
	} else {
		width = base.Dx()
		height = base.Dy() / n
	}
	if width == 0 || height == 0 {
		return
	}
	for i := range n {
		var r image.Rectangle
		if mode == SplitHorizontalMode {
			r = image.Rect(
				base.Min.X+width*i, base.Min.Y,
				base.Min.X+width*(i+1), base.Min.Y+height,
			)
		} else {
			r = image.Rect(
				base.Min.X, base.Min.Y+height*i,
				base.Min.X+width, base.Min.Y+height*(i+1),
			)
		}
		rects = append(rects, r)
	}
	return
}

// Split splits an image into n smaller images based on the specified split mode.
// If n is less than 1, or the image cannot be split, it returns an error.
func Split(base image.Image, n int, mode SplitMode) (imgs []image.Image, err error) {
	if n < 1 {
		return nil, errors.New("invalid number of parts: must be at least 1")
	}
	if img, ok := base.(interface {
		SubImage(image.Rectangle) image.Image
	}); ok {
		rects := split(base.Bounds(), n, mode)
		if len(rects) == 0 {
			return nil, errors.New("failed to split the image: invalid dimensions or n is too large")
		}
		for _, rect := range rects {
			imgs = append(imgs, img.SubImage(rect))
		}
	} else {
		return nil, errors.New("image type does not support SubImage extraction")
	}
	return
}

// SplitHorizontal splits an image into n parts horizontally.
func SplitHorizontal(base image.Image, n int) ([]image.Image, error) {
	return Split(base, n, SplitHorizontalMode)
}

// SplitVertical splits an image into n parts vertically.
func SplitVertical(base image.Image, n int) ([]image.Image, error) {
	return Split(base, n, SplitVerticalMode)
}
