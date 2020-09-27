package main

import (
	"flag"
	"fmt"
	"image"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/sunshineplan/img"
	"github.com/sunshineplan/utils/workers"
	"github.com/vharitonsky/iniflags"
)

var self string
var src, dst string
var quality int
var watermark string
var opacity uint
var random bool
var offsetX, offsetY int
var width, height int
var percent float64
var debug bool

func init() {
	var err error
	self, err = os.Executable()
	if err != nil {
		log.Fatalf("Failed to get self path: %v", err)
	}
}

func usage() {
	fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
	fmt.Println(`
  --src
		source file or directory
  --dst
		destination file or directory
  --quality
		set output file quality (range 1-100, default: 75)
  --watermark
		watermark name (default: watermark.png)
  --opacity
		watermark opacity (range 0-255, default: 128)
  --random
		random watermark
  -x, y
		fixed watermark center offset X, Y value. Only used in no random mode.
  --width
		resize width, if one of width or height is 0, the image aspect ratio is preserved.
  --height
		resize height, if one of width or height is 0, the image aspect ratio is preserved.
  --percent
		resize percent, only when both of width and height are 0.`)
}

func main() {
	flag.Usage = usage
	flag.StringVar(&src, "src", "", "")
	flag.StringVar(&dst, "dst", "", "")
	flag.IntVar(&quality, "quality", 75, "")
	flag.StringVar(&watermark, "watermark", "", "")
	flag.UintVar(&opacity, "opacity", 128, "")
	flag.BoolVar(&random, "random", false, "")
	flag.IntVar(&offsetX, "x", 0, "")
	flag.IntVar(&offsetY, "y", 0, "")
	flag.IntVar(&width, "width", 0, "")
	flag.IntVar(&height, "height", 0, "")
	flag.Float64Var(&percent, "percent", 0, "")
	flag.BoolVar(&debug, "debug", false, "")
	iniflags.SetConfigFile(filepath.Join(filepath.Dir(self), "config.ini"))
	iniflags.SetAllowMissingConfigFile(true)
	iniflags.Parse()

	f, err := os.OpenFile(filepath.Join(filepath.Dir(self), "convert.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0640)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer f.Close()
	log.SetOutput(io.MultiWriter(f, os.Stdout))

	task := img.Option{Quality: quality}
	if watermark != "" {
		task.SetWatermark(watermark, opacity, random, image.Point{X: offsetX, Y: offsetY})
	}
	if width != 0 || height != 0 || percent != 0 {
		task.SetResize(width, height, percent)
	}

	if !task.Test() {
		log.Fatal("No task could be found.")
	}

	si, err := os.Stat(src)
	if err != nil {
		log.Fatal(err)
	}
	di, err := os.Stat(dst)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(dst, 0755); err != nil {
			log.Fatal(err)
		}
		di, _ = os.Stat(dst)
	} else if err != nil {
		log.Fatal(err)
	}
	switch mode := si.Mode(); {
	case mode.IsDir():
		if di.Mode().IsDir() {
			var images []string
			filepath.Walk(src, func(path string, _ os.FileInfo, _ error) error {
				if ok, _ := regexp.MatchString(`^\.(jpe?g|png|gif|tiff?|bmp)$`, strings.ToLower(filepath.Ext(path))); ok {
					images = append(images, path)
				}
				return nil
			})
			workers.Slice(images, func(_ int, i interface{}) {
				var output string
				if debug {
					log.Printf("Converting %s to %s\n", i.(string), output)
				}
				if err := task.Convert(i.(string), output); err != nil {
					log.Print(err)
				}
			})
		} else {
			log.Fatal("Destination is not a directory.")
		}
	case mode.IsRegular():
		var output string
		switch mode := di.Mode(); {
		case mode.IsDir():
			output = filepath.Join(dst, filepath.Base(src))
		case mode.IsRegular():
			output = dst
		}
		if err := task.Convert(src, output); err != nil {
			log.Fatal(err)
		}
	}
}
