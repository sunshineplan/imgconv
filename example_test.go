package imgconv_test

import (
	"io/ioutil"
	"log"

	"github.com/sunshineplan/imgconv"
)

func Example() {
	// Open a test image.
	src, err := imgconv.Open("testdata/video-001.png")
	if err != nil {
		log.Fatalf("failed to open image: %v", err)
	}

	// Resize the cropped image to width = 200px preserving the aspect ratio.
	mark := imgconv.Resize(src, imgconv.ResizeOption{Percent: 25})

	// Create a blurred version of the image.
	dst := imgconv.Watermark(src, imgconv.WatermarkOption{Mark: mark, Opacity: 128, Random: true})

	// Save the resulting image as JPEG.
	task := imgconv.New()
	task.SetFormat("jpg")
	err = task.Convert(dst, ioutil.Discard)
	if err != nil {
		log.Fatalf("failed to save image: %v", err)
	}
}
