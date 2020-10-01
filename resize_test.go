package imgconv

import (
	"image"
	"testing"
)

func TestResize(t *testing.T) {
	testCase := []struct {
		option ResizeOption
		want   image.Point
	}{
		{ResizeOption{Width: 300}, image.Pt(300, 206)},
		{ResizeOption{Height: 206}, image.Pt(300, 206)},
		{ResizeOption{Width: 200, Height: 200}, image.Pt(200, 200)},
		{ResizeOption{Percent: 50}, image.Pt(75, 52)},
	}

	// Read the image.
	sample, err := Open("testdata/video-001.png")
	if err != nil {
		t.Error(err)
		return
	}
	for _, tc := range testCase {
		// Resize the image.
		got := tc.option.do(sample)
		if got.Bounds().Size() != tc.want {
			t.Errorf("bounds differ: %v and %v", got.Bounds().Size(), tc.want)
			continue
		}
	}
}
