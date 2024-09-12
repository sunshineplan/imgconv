package imgconv

import (
	"image"
	"slices"
	"testing"
)

func TestSplit(t *testing.T) {
	for i, testcase := range []struct {
		base image.Rectangle
		n    int
		mode SplitMode
		want []image.Rectangle
	}{
		{image.Rect(0, 0, 100, 100), 4, SplitHorizontalMode, []image.Rectangle{
			image.Rect(0, 0, 25, 100),
			image.Rect(25, 0, 50, 100),
			image.Rect(50, 0, 75, 100),
			image.Rect(75, 0, 100, 100),
		}},
		{image.Rect(0, 0, 100, 100), 4, SplitVerticalMode, []image.Rectangle{
			image.Rect(0, 0, 100, 25),
			image.Rect(0, 25, 100, 50),
			image.Rect(0, 50, 100, 75),
			image.Rect(0, 75, 100, 100),
		}},
		{image.Rect(100, 100, 200, 200), 4, SplitHorizontalMode, []image.Rectangle{
			image.Rect(100, 100, 125, 200),
			image.Rect(125, 100, 150, 200),
			image.Rect(150, 100, 175, 200),
			image.Rect(175, 100, 200, 200),
		}},
		{image.Rect(100, 100, 200, 200), 4, SplitVerticalMode, []image.Rectangle{
			image.Rect(100, 100, 200, 125),
			image.Rect(100, 125, 200, 150),
			image.Rect(100, 150, 200, 175),
			image.Rect(100, 175, 200, 200),
		}},
	} {
		if rects := split(testcase.base, testcase.n, testcase.mode); slices.CompareFunc(
			rects,
			testcase.want,
			func(a, b image.Rectangle) int {
				if a.Eq(b) {
					return 0
				}
				return 1
			},
		) != 0 {
			t.Errorf("#%d wrong split results: want %v, got %v", i, testcase.want, rects)
		}
	}
}

func TestSplitError(t *testing.T) {
	r := image.Rect(0, 0, 100, 100)
	img := image.NewNRGBA(r)
	if _, err := Split(img, 10, SplitHorizontalMode); err != nil {
		t.Fatal(err)
	}
	for i, testcase := range []struct {
		img image.Image
		n   int
	}{
		{r, 10},
		{img, 0},
		{img, 101},
		{image.NewNRGBA(image.Rectangle{}), 10},
	} {
		if _, err := Split(testcase.img, testcase.n, SplitHorizontalMode); err == nil {
			t.Errorf("#%d want error, got nil", i)
		}
	}
}
