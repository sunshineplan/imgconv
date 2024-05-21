package imgconv

import (
	"image"
	"reflect"
	"slices"
	"testing"
	"time"
)

func TestWatermark(t *testing.T) {
	mark := &image.NRGBA{
		Rect:   image.Rect(0, 0, 4, 4),
		Stride: 4 * 4,
		Pix: []uint8{
			0x00, 0x00, 0x00, 0x00, 0xff, 0x00, 0x00, 0x40, 0xff, 0x00, 0x00, 0xbf, 0xff, 0x00, 0x00, 0xff,
			0x00, 0xff, 0x00, 0x40, 0x6e, 0x6d, 0x25, 0x70, 0xb0, 0x14, 0x3b, 0xcf, 0xbf, 0x00, 0x40, 0xff,
			0x00, 0xff, 0x00, 0xbf, 0x14, 0xb0, 0x3b, 0xcf, 0x33, 0x33, 0x99, 0xef, 0x40, 0x00, 0xbf, 0xff,
			0x00, 0xff, 0x00, 0xff, 0x00, 0xbf, 0x40, 0xff, 0x00, 0x40, 0xbf, 0xff, 0x00, 0x00, 0xff, 0xff,
		},
	}

	// Read the image.
	sample, err := Open("testdata/video-001.png")
	if err != nil {
		t.Fatal("testdata/video-001.png", err)
	}

	m0 := (&WatermarkOption{Mark: mark, Opacity: 50}).SetOffset(image.Pt(5, 5)).do(sample)
	m1 := Watermark(sample, &WatermarkOption{Mark: mark, Opacity: 50, Offset: image.Pt(5, 5)})
	if !reflect.DeepEqual(m0, m1) {
		t.Fatal("Fixed Watermark got different images")
	}

	m0 = (&WatermarkOption{Mark: mark, Opacity: 50}).SetRandom(true).do(sample)
	time.Sleep(time.Nanosecond)
	m1 = (&WatermarkOption{Mark: mark, Opacity: 50, Random: true}).do(sample)
	if reflect.DeepEqual(m0, m1) {
		t.Fatal("Random Watermark got same images")
	}

	(&WatermarkOption{Mark: sample, Random: true}).do(sample)
	(&WatermarkOption{Mark: sample, Random: true}).do(rotate90(sample))
}

func TestCalcResizeXY(t *testing.T) {
	testCase := []struct {
		base image.Rectangle
		mark image.Rectangle
		want bool
	}{
		{image.Rect(0, 0, 100, 50), image.Rect(0, 0, 200, 200), false},
		{image.Rect(0, 0, 50, 100), image.Rect(0, 0, 200, 200), true},
	}

	for _, tc := range testCase {
		if calcResizeXY(tc.base, tc.mark) != tc.want {
			t.Errorf("Want %v, got %v", tc.want, !tc.want)
		}
	}
}

func TestRandRange(t *testing.T) {
	testCase := []struct {
		min, max int
		res      []int
	}{
		{1, 5, []int{1, 2, 3, 4, 5}},
		{5, 1, []int{1, 2, 3, 4, 5}},
		{-1, 5, []int{-1, 0, 1, 2, 3, 4, 5}},
		{5, -1, []int{-1, 0, 1, 2, 3, 4, 5}},
		{-5, -1, []int{-5, -4, -3, -2, -1}},
		{-1, -5, []int{-5, -4, -3, -2, -1}},
	}

	for i := 0; i < 100; i++ {
		for i, tc := range testCase {
			if res := randRange(tc.min, tc.max); !slices.Contains(tc.res, res) {
				t.Errorf("#%d: got %d, not in range %v", i, res, tc.res)
			}
		}
	}
}
