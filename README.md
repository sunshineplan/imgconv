# Image Converter

[![GoDev](https://img.shields.io/static/v1?label=godev&message=reference&color=00add8)][godev]
[![BuildStatus](https://travis-ci.org/sunshineplan/imgconv.svg?branch=master)][travis]
[![CoverageStatus](https://coveralls.io/repos/github/sunshineplan/imgconv/badge.svg?branch=master&service=github)][coveralls]
[![GoReportCard](https://goreportcard.com/badge/github.com/sunshineplan/imgconv)][goreportcard]

[godev]: https://pkg.go.dev/github.com/sunshineplan/imgconv
[travis]: https://travis-ci.org/sunshineplan/imgconv
[coveralls]: https://coveralls.io/github/sunshineplan/imgconv?branch=master
[goreportcard]: https://goreportcard.com/report/github.com/sunshineplan/imgconv

Package imgconv provides basic image processing functions (resize, add watermark, format converter.).

All the image processing functions provided by the package accept any image type that implements `image.Image` interface
as an input, and return a new image of `*image.NRGBA` type (32bit RGBA colors, non-premultiplied alpha).

## Installation

    go get -u github.com/sunshineplan/imgconv

## Documentation

https://pkg.go.dev/github.com/sunshineplan/imgconv

## License

[The MIT License (MIT)](https://raw.githubusercontent.com/sunshineplan/imgconv/master/LICENSE)

## Credits

This repo relies on the following third-party projects:

  * [disintegration/imaging](https://github.com/disintegration/imaging)

## Usage examples

A few usage examples can be found below. See the documentation for the full list of supported functions.

### Image resizing

```go
// Resize srcImage to size = 128x128px.
dstImage128 := imgconv.Resize(srcImage, imgconv.ResizeOption{Width: 128, Height: 128})

// Resize srcImage to width = 800px preserving the aspect ratio.
dstImage800 := imgconv.Resize(srcImage, imgconv.ResizeOption{Width: 800})

// Resize srcImage to 50% size preserving the aspect ratio.
dstImagePercent50 := imgconv.Resize(srcImage, imgconv.ResizeOption{Percent: 50})
```

### Add watermark

```go
// srcImage add a watermark at randomly position.
dstImage := imgconv.Watermark(srcImage, WatermarkOption{Mark: markImage, Opacity: 128, Random: true})

// srcImage add a watermark at fixed position with offset.
dstImage := imgconv.Watermark(srcImage, WatermarkOption{Mark: markImage, Opacity: 128, Offset: image.Pt(5, 5)})
```

### Format convert

```go
// Convert srcImage to dst with jpg format.
imgconv.Write(srcImage, dstWriter, imgconv.FormatOption{Format: imgconv.JPEG})
```

## Example code

```go
package main

import (
	"io/ioutil"
	"log"

	"github.com/sunshineplan/imgconv"
)

func main() {
	// Open a test image.
	src, err := imgconv.Open("testdata/video-001.png")
	if err != nil {
		log.Fatalf("failed to open image: %v", err)
	}

	// Resize the image to width = 200px preserving the aspect ratio.
	mark := imgconv.Resize(src, imgconv.ResizeOption{Width: 200})

	// Add random watermark set opacity = 128.
	dst := imgconv.Watermark(src, imgconv.WatermarkOption{Mark: mark, Opacity: 128, Random: true})

	// Write the resulting image as TIFF.
	err = imgconv.Write(dst, ioutil.Discard, imgconv.FormatOption{Format: imgconv.TIFF})
	if err != nil {
		log.Fatalf("failed to write image: %v", err)
	}
}
```
