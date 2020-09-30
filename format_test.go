package imgconv

import (
	"bytes"
	"image"
	"image/png"
	"os"
	"testing"
)

func readPng(filename string) (image.Image, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return png.Decode(f)
}

func TestEncode(t *testing.T) {
	testCase := []FormatOption{
		{Format: JPEG, EncodeOption: []EncodeOption{JPEGQuality(75)}},
		{Format: PNG, EncodeOption: []EncodeOption{PNGCompressionLevel(png.DefaultCompression)}},
		{Format: GIF, EncodeOption: []EncodeOption{GIFNumColors(256)}},
		{Format: TIFF},
		{Format: BMP},
	}

	// Read the image.
	m0, err := readPng("testdata/video-001.png")
	if err != nil {
		t.Error("testdata/video-001.png", err)
		return
	}
	for _, tc := range testCase {
		// Encode the image.
		var buf bytes.Buffer
		if err := Write(m0, &buf, tc); err != nil {
			t.Error(formatExts[tc.Format], err)
			continue
		}
		// Decode the image.
		m1, err := decode(&buf, tc.Format)
		if err != nil {
			t.Error(formatExts[tc.Format], err)
			continue
		}
		if m0.Bounds() != m1.Bounds() {
			t.Errorf("bounds differ: %v and %v", m0.Bounds(), m1.Bounds())
			continue
		}
	}
}
