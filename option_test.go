package imgconv

import (
	"bytes"
	"image"
	"io"
	"testing"
)

func TestOption(t *testing.T) {
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

	opts := NewOptions()
	if opts.Format.Format != JPEG {
		t.Fatal("Format is not expect one.")
	}
	opts.SetWatermark(mark, 100)
	if mark != opts.Watermark.Mark || opts.Watermark.Opacity != 100 {
		t.Fatal("SetWatermark result is not expect one.")
	}
	opts.SetWatermark(mark, 0)
	if mark != opts.Watermark.Mark || opts.Watermark.Opacity != 128 {
		t.Fatal("SetWatermark result is not expect one.")
	}
	opts.SetResize(0, 0, 33)
	if opts.Resize.Width != 0 || opts.Resize.Height != 0 || opts.Resize.Percent != 33 {
		t.Fatal("SetResize result is not expect one.")
	}
	opts.SetGray(true)
	if !opts.Gray {
		t.Fatal("SetGray result is not expect one.")
	}
	if err := opts.Convert(io.Discard, mark); err != nil {
		t.Fatal("Failed to Convert.")
	}
}

func TestConvert(t *testing.T) {
	base, err := Open("testdata/video-001.png")
	if err != nil {
		t.Fatal(err)
	}

	var buf1, buf2 bytes.Buffer
	opts := NewOptions()
	if err := opts.Convert(&buf1, base); err != nil {
		t.Fatal("Failed to Convert.")
	}
	opts = &Options{Format: &FormatOption{}}
	if err := opts.Convert(&buf2, base); err != nil {
		t.Fatal("Failed to Convert.")
	}

	if !bytes.Equal(buf1.Bytes(), buf2.Bytes()) {
		t.Fatal("Convert get different result")
	}
}

func TestConvertExt(t *testing.T) {
	opts := NewOptions().SetFormat(TIFF)
	if opts.ConvertExt("testdata/video-001.png") != "testdata/video-001.tif" {
		t.Fatal("ConvertExt result is not expect one.")
	}
}
