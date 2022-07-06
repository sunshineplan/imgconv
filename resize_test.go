package imgconv

import (
	"image"
	"testing"
)

func compare(t *testing.T, img0, img1 image.Image) {
	t.Helper()
	b0 := img0.Bounds()
	b1 := img1.Bounds()
	if b0.Dx() != b1.Dx() || b0.Dy() != b1.Dy() {
		t.Fatalf("wrong image size: want %s, got %s", b0, b1)
	}
	x1 := b1.Min.X - b0.Min.X
	y1 := b1.Min.Y - b0.Min.Y
	for y := b0.Min.Y; y < b0.Max.Y; y++ {
		for x := b0.Min.X; x < b0.Max.X; x++ {
			c0 := img0.At(x, y)
			c1 := img1.At(x+x1, y+y1)
			r0, g0, b0, a0 := c0.RGBA()
			r1, g1, b1, a1 := c1.RGBA()
			if r0 != r1 || g0 != g1 || b0 != b1 || a0 != a1 {
				t.Fatalf("pixel at (%d, %d) has wrong color: want %v, got %v", x, y, c0, c1)
			}
		}
	}
}

func TestResize(t *testing.T) {
	testCase := []struct {
		option *ResizeOption
		want   image.Point
	}{
		{&ResizeOption{Width: 300}, image.Pt(300, 206)},
		{&ResizeOption{Height: 206}, image.Pt(300, 206)},
		{&ResizeOption{Width: 200, Height: 200}, image.Pt(200, 200)},
		{&ResizeOption{Percent: 50}, image.Pt(75, 52)},
	}

	// Read the image.
	sample, err := Open("testdata/video-001.png")
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range testCase {
		// Resize the image.
		img0 := tc.option.do(sample)
		if img0.Bounds().Size() != tc.want {
			t.Fatalf("bounds differ: %v and %v", img0.Bounds().Size(), tc.want)
		}
		img1 := Resize(sample, tc.option)

		compare(t, img0, img1)
	}
}
