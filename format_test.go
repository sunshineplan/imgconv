package imgconv

import (
	"bytes"
	"image"
	"image/png"
	"os"
	"testing"

	"github.com/disintegration/imaging"
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
	testCase := []formatOption{
		{format: imaging.JPEG, encodeOption: []imaging.EncodeOption{imaging.JPEGQuality(75)}},
		{format: imaging.PNG, encodeOption: []imaging.EncodeOption{imaging.PNGCompressionLevel(png.DefaultCompression)}},
		{format: imaging.GIF, encodeOption: []imaging.EncodeOption{imaging.GIFNumColors(256)}},
		{format: imaging.TIFF},
		{format: imaging.BMP},
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
		if err := Encode(m0, &buf, tc); err != nil {
			t.Error(tc.format.String(), err)
			continue
		}
		// Decode the image.
		m1, err := decode(&buf, tc.format)
		if err != nil {
			t.Error(tc.format.String(), err)
			continue
		}
		if m0.Bounds() != m1.Bounds() {
			t.Errorf("bounds differ: %v and %v", m0.Bounds(), m1.Bounds())
			continue
		}
	}
}
