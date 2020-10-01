package imgconv

import (
	"bytes"
	"image/draw"
	"image/png"
	"io/ioutil"
	"reflect"
	"testing"
)

func TestEncode(t *testing.T) {
	testCase := []FormatOption{
		{Format: JPEG, EncodeOption: []EncodeOption{JPEGQuality(75)}},
		{Format: PNG, EncodeOption: []EncodeOption{PNGCompressionLevel(png.DefaultCompression)}},
		{Format: GIF, EncodeOption: []EncodeOption{GIFNumColors(256), GIFDrawer(draw.FloydSteinberg), GIFQuantizer(nil)}},
		{Format: TIFF},
		{Format: BMP},
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
		if err := Write(m0, &buf, fo); err != nil {
			t.Error(formatExts[fo.Format], err)
			continue
		}
		// Decode the image.
		b, _ := ioutil.ReadAll(&buf)
		m1, err := decode(bytes.NewBuffer(b), fo.Format)
		if err != nil {
			t.Error(formatExts[fo.Format], err)
			continue
		}
		m2, err := Decode(bytes.NewBuffer(b))
		if err != nil {
			t.Error(err)
			continue
		}
		if !reflect.DeepEqual(m1, m2) {
			t.Error("Decode get different images")
		}
		if m0.Bounds() != m1.Bounds() {
			t.Errorf("bounds differ: %v and %v", m0.Bounds(), m1.Bounds())
			continue
		}
	}
}
