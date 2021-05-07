package imgconv

import (
	"bytes"
	"image"
	"image/draw"
	"image/png"
	"io"
	"testing"

	"github.com/sunshineplan/tiff"
)

func TestSetFormat(t *testing.T) {
	if _, err := setFormat("Jpg"); err != nil {
		t.Error("Failed to set format")
	}
	if _, err := setFormat("txt"); err == nil {
		t.Error("set txt format want error")
	}
}

func TestEncode(t *testing.T) {
	testCase := []FormatOption{
		{Format: JPEG, EncodeOption: []EncodeOption{Quality(75)}},
		{Format: PNG, EncodeOption: []EncodeOption{PNGCompressionLevel(png.DefaultCompression)}},
		{Format: GIF, EncodeOption: []EncodeOption{GIFNumColors(256), GIFDrawer(draw.FloydSteinberg), GIFQuantizer(nil)}},
		{Format: TIFF, EncodeOption: []EncodeOption{TIFFCompressionType(tiff.LZW)}},
		{Format: BMP},
		{Format: PDF, EncodeOption: []EncodeOption{Quality(75)}},
	}

	// Read the image.
	m0, err := Open("testdata/video-001.png")
	if err != nil {
		t.Error(err)
		return
	}
	for _, tc := range testCase {
		// Encode the image.
		var buf bytes.Buffer
		fo, err := setFormat(formatExts[tc.Format], tc.EncodeOption...)
		if err != nil {
			t.Error(tc, err)
			continue
		}
		if err := fo.Encode(&buf, m0); err != nil {
			t.Error(formatExts[fo.Format], err)
			continue
		}
		// Decode the image.
		m1, err := Decode(&buf)
		if err != nil {
			t.Error(formatExts[fo.Format], err)
			continue
		}
		if m0.Bounds() != m1.Bounds() {
			t.Errorf("bounds differ: %v and %v", m0.Bounds(), m1.Bounds())
			continue
		}
	}
	if err := (&FormatOption{}).Encode(io.Discard, &image.NRGBA{
		Rect:   image.Rect(0, 0, 1, 1),
		Stride: 1 * 4,
		Pix:    []uint8{0xff, 0xff, 0xff, 0xff}}); err != nil {
		t.Error("encode image error")
	}
	if err := (&FormatOption{Format: -1}).Encode(io.Discard, m0); err == nil {
		t.Error("encode unsupported format expect an error")
	}
}
