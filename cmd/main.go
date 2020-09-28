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
	"time"

	"github.com/disintegration/imaging"
	"github.com/sunshineplan/img"
	"github.com/sunshineplan/utils/workers"
	"github.com/vharitonsky/iniflags"
)

var self string
var src, dst string
var format string
var quality int
var watermark string
var opacity uint
var random bool
var offsetX, offsetY int
var width, height int
var percent float64
var debug bool
var worker int

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
		destination directory (default: output)
  --format
		output format (jpg, jpeg, png, gif, tif, tiff and bmp are supported, default: jpg)
  --quality
		set jpeg quality (range 1-100, default: 75)
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
	flag.StringVar(&dst, "dst", "output", "")
	flag.StringVar(&format, "format", "jpg", "")
	flag.IntVar(&quality, "quality", 75, "")
	flag.StringVar(&watermark, "watermark", "", "")
	flag.UintVar(&opacity, "opacity", 128, "")
	flag.BoolVar(&random, "random", false, "")
	flag.IntVar(&offsetX, "x", 0, "")
	flag.IntVar(&offsetY, "y", 0, "")
	flag.IntVar(&width, "width", 0, "")
	flag.IntVar(&height, "height", 0, "")
	flag.Float64Var(&percent, "percent", 0, "")
	flag.IntVar(&worker, "worker", 5, "")
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

	task := img.New()
	if outputFormat, err := imaging.FormatFromExtension(format); err == nil {
		task.SetFormat(outputFormat, imaging.JPEGQuality(quality))
	} else {
		log.Fatalln("Unknown output format:", format)
	}
	if watermark != "" {
		task.SetWatermark(watermark, opacity, random, image.Point{X: offsetX, Y: offsetY})
	}
	if width != 0 || height != 0 || percent != 0 {
		task.SetResize(width, height, percent)
	}

	si, err := os.Stat(src)
	if err != nil {
		log.Fatal(err)
	}
	di, err := os.Stat(dst)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(dst, 0755); err != nil {
				log.Fatal(err)
			}
			di, _ = os.Stat(dst)
		} else {
			log.Fatal(err)
		}
	} else if !di.Mode().IsDir() {
		log.Fatal("Destination is not a directory.")
	}
	switch mode := si.Mode(); {
	case mode.IsDir():
		var images []string
		filepath.Walk(src, func(path string, _ os.FileInfo, _ error) error {
			if ok, _ := regexp.MatchString(`^\.(jpe?g|png|gif|tiff?|bmp)$`, strings.ToLower(filepath.Ext(path))); ok {
				images = append(images, path)
			}
			return nil
		})
		total := len(images)
		log.Println("Total images:", total)
		start := time.Now()
		var count int
		go func() {
			for {
				done := 50 * count / total
				speed := float64(count) / float64(time.Since(start)/time.Second)
				percent := float64(count) * 100 / float64(total)
				left := time.Duration(float64(total-count)/speed) * time.Second
				fmt.Printf("   [%s%s]   %.2f/s - %d(%.2f%%) of %d   Left: %v     \r",
					strings.Repeat("=", done), strings.Repeat(" ", 50-done), speed, count, percent, total, left)
				time.Sleep(time.Second)
			}
		}()
		workers.New(worker).Slice(images, func(_ int, i interface{}) {
			rel, _ := filepath.Rel(src, i.(string))
			output := filepath.Join(dst, rel)
			if err := task.Convert(i.(string), output); err != nil {
				if err == os.ErrExist {
					log.Println("Skip", i.(string))
				} else {
					log.Println(i, err)
				}
				count++
				return
			}
			if debug {
				log.Printf("[Debug]Converted %s\n", i.(string))
			}
			count++
		})
		log.Println("Job done! Elapsed time:", time.Since(start))
	case mode.IsRegular():
		filename := filepath.Base(src)
		if err := task.Convert(src, filepath.Join(dst, filename)); err != nil {
			log.Fatal(err)
		}
	default:
		log.Fatal("Unknown source.")
	}
}
