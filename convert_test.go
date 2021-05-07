package imgconv

import (
	"bytes"
	"io"
	"os"
	"testing"
)

func TestDecodeWrite(t *testing.T) {
	var format = []string{
		"jpg",
		"png",
		"gif",
		"tif",
		"bmp",
		"webp",
		"pdf",
	}
	for _, i := range format {
		b, err := os.ReadFile("testdata/video-001." + i)
		if err != nil {
			t.Error(err)
			continue
		}
		img, err := Decode(bytes.NewBuffer(b))
		if err != nil {
			t.Error("Failed to decode", i)
		}
		if err := Write(io.Discard, img, FormatOption{}); err != nil {
			t.Error("Failed to write", i)
		}
		if _, err := DecodeConfig(bytes.NewBuffer(b)); err != nil {
			t.Error("Failed to decode", i, "config")
		}
	}
	if _, err := Decode(bytes.NewBufferString("Hello")); err == nil {
		t.Error("Decode string want error")
	}
}

func TestOpenSave(t *testing.T) {
	if _, err := Open("/invalid/path"); err == nil {
		t.Error("Open invalid path want error")
	}
	if _, err := Open("build.bat"); err == nil {
		t.Error("Open invalid image want error")
	}
	img, err := Open("testdata/video-001.png")
	if err != nil {
		t.Error("Fail to open image", err)
		return
	}
	if err := Save("/invalid/path", img, defaultFormat); err == nil {
		t.Error("Save invalid path want error")
	}
	if err := Save("testdata/tmp", img, defaultFormat); err != nil {
		t.Error("Fail to save image", err)
		return
	}
	if err := os.Remove("testdata/tmp"); err != nil {
		t.Error(err)
	}
}
