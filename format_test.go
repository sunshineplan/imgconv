package imgconv

import (
	"bytes"
	"flag"
	"image"
	"image/draw"
	"image/png"
	"io"
	"testing"
)

func TestFormatFromExtension(t *testing.T) {
	if _, err := FormatFromExtension("Jpg"); err != nil {
		t.Fatal("jpg format want no error")
	}
	if _, err := FormatFromExtension("TIFF"); err != nil {
		t.Fatal("tiff format want no error")
	}
	if _, err := FormatFromExtension("txt"); err == nil {
		t.Fatal("txt format want error")
	}
}

func TestTextVar(t *testing.T) {
	testCase1 := []struct {
		argument string
		format   Format
	}{
		{"Jpg", JPEG},
		{"TIFF", TIFF},
		{"txt", Format(-1)},
	}
	for _, tc := range testCase1 {
		f := flag.NewFlagSet("test", flag.ContinueOnError)
		f.SetOutput(io.Discard)
		var format Format
		f.TextVar(&format, "f", Format(-1), "")
		f.Parse(append([]string{"-f"}, tc.argument))
		if format != tc.format {
			t.Errorf("expected %s format; got %s", tc.format, format)
		}
	}
	testCase2 := []struct {
		argument    string
		compression TIFFCompression
	}{
		{"none", TIFFUncompressed},
		{"Deflate", TIFFDeflate},
		{"lzw", TIFFCompression(-1)},
	}
	for _, tc := range testCase2 {
		f := flag.NewFlagSet("test", flag.ContinueOnError)
		f.SetOutput(io.Discard)
		var compression TIFFCompression
		f.TextVar(&compression, "c", TIFFCompression(-1), "")
		f.Parse(append([]string{"-c"}, tc.argument))
		if compression != tc.compression {
			t.Errorf("expected %d compression; got %d", tc.compression, compression)
		}
	}
}

func TestEncode(t *testing.T) {
	testCase := []FormatOption{
		{Format: JPEG, EncodeOption: []EncodeOption{Quality(75)}},
		{Format: PNG, EncodeOption: []EncodeOption{PNGCompressionLevel(png.DefaultCompression)}},
		{Format: GIF, EncodeOption: []EncodeOption{GIFNumColors(256), GIFDrawer(draw.FloydSteinberg), GIFQuantizer(nil)}},
		{Format: TIFF, EncodeOption: []EncodeOption{TIFFCompressionType(TIFFDeflate)}},
		{Format: BMP},
		{Format: PDF, EncodeOption: []EncodeOption{Quality(75)}},
	}

	// Read the image.
	m0, err := Open("testdata/video-001.png")
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range testCase {
		// Encode the image.
		var buf bytes.Buffer
		fo := &FormatOption{tc.Format, tc.EncodeOption}
		if err := fo.Encode(&buf, m0); err != nil {
			t.Fatal(formatExts[fo.Format], err)
		}

		// Decode the image.
		m1, err := Decode(&buf)
		if err != nil {
			t.Fatal(formatExts[fo.Format], err)
		}

		if m0.Bounds() != m1.Bounds() {
			t.Fatalf("bounds differ: %v and %v", m0.Bounds(), m1.Bounds())
		}
	}

	if err := (&FormatOption{}).Encode(
		io.Discard,
		&image.NRGBA{
			Rect:   image.Rect(0, 0, 1, 1),
			Stride: 1 * 4,
			Pix:    []uint8{0xff, 0xff, 0xff, 0xff},
		},
	); err != nil {
		t.Fatal("encode image error")
	}

	if err := (&FormatOption{Format: -1}).Encode(io.Discard, m0); err == nil {
		t.Fatal("encode unsupported format expect an error")
	}
}
