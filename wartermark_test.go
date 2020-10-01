package imgconv

import (
	"image"
	"reflect"
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
		t.Error("testdata/video-001.png", err)
		return
	}

	m0 := (&WatermarkOption{Mark: mark, Opacity: 50}).SetOffset(image.Pt(5, 5)).do(sample)
	m1 := (&WatermarkOption{Mark: mark, Opacity: 50, Offset: image.Pt(5, 5)}).do(sample)
	if !reflect.DeepEqual(m0, m1) {
		t.Error("Fixed Watermark get different images")
	}
	m0 = (&WatermarkOption{Mark: mark, Opacity: 50}).SetRandom(true).do(sample)
	time.Sleep(time.Nanosecond)
	m1 = (&WatermarkOption{Mark: mark, Opacity: 50, Random: true}).do(sample)
	if reflect.DeepEqual(m0, m1) {
		t.Error("Random Watermark get same image")
	}
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
			t.Errorf("Want %v, get %v", tc.want, !tc.want)
		}
	}
}
