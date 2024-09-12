package imgconv_test

import (
	"fmt"
	"io"
	"log"

	"github.com/sunshineplan/imgconv"
)

func Example() {
	// Open a test image.
	src, err := imgconv.Open("testdata/video-001.png")
	if err != nil {
		log.Fatalf("failed to open image: %v", err)
	}

	// Resize the image to width = 200px preserving the aspect ratio.
	mark := imgconv.Resize(src, &imgconv.ResizeOption{Width: 200})

	// Add random watermark set opacity = 128.
	dst := imgconv.Watermark(src, &imgconv.WatermarkOption{Mark: mark, Opacity: 128, Random: true})

	// Write the resulting image as TIFF.
	if err := imgconv.Write(io.Discard, dst, &imgconv.FormatOption{Format: imgconv.TIFF}); err != nil {
		log.Fatalf("failed to write image: %v", err)
	}

	// Split the image into 3 parts horizontally.
	imgs, err := imgconv.Split(src, 3, imgconv.SplitHorizontalMode)
	if err != nil {
		log.Fatalf("failed to split image: %v", err)
	}
	fmt.Print(len(imgs))
	// output:3
}
