package imgconv

import (
	"image"
	"testing"
)

func TestGray(t *testing.T) {
	sample, err := Open("testdata/video-001.png")
	if err != nil {
		t.Fatal(err)
	}

	img := ToGray(sample)
	if img.Bounds().Size() != sample.Bounds().Size() {
		t.Fatalf("bounds differ: %v and %v", img.Bounds().Size(), sample.Bounds().Size())
	}
	if _, ok := img.(*image.Gray); !ok {
		t.Fatal("img is not gray")
	}
}
