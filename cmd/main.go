package main

import (
	"flag"
	"fmt"
	"image"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/sunshineplan/imgconv"
	"github.com/sunshineplan/tiff"
	"github.com/sunshineplan/utils/progressbar"
	"github.com/sunshineplan/utils/workers"
	"github.com/vharitonsky/iniflags"
)

var self string
var src, dst string
var format string
var quality int
var compression string
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
		output format (jpg, jpeg, png, gif, tif, tiff, bmp and pdf are supported, default: jpg)
  --quality
		set jpeg or pdf quality (range 1-100, default: 75)
  --compression
		set tiff compression type (none, lzw, deflate, default: lzw)
  --watermark
		watermark path
  --opacity
		watermark opacity (range 0-255, default: 128)
  --random
		random watermark (default: false)
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
	flag.StringVar(&compression, "compression", "lzw", "")
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

	f, err := os.OpenFile(
		filepath.Join(filepath.Dir(self), fmt.Sprintf("convert%s.log", time.Now().Format("20060102150405"))),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0640)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer f.Close()

	log.SetOutput(io.MultiWriter(f, os.Stdout))

	task := imgconv.New()
	if err != nil {
		log.Fatal(err)
	}

	var ct tiff.CompressionType
	switch strings.ToLower(compression) {
	case "none":
		ct = tiff.Uncompressed
	case "lzw":
		ct = tiff.LZW
	case "deflate":
		ct = tiff.Deflate
	default:
		log.Fatalln("Unknown tiff compression type:", ct)
	}

	if err := task.SetFormat(format, imgconv.Quality(quality), imgconv.TIFFCompressionType(ct)); err != nil {
		log.Fatal(err)
	}

	if watermark != "" {
		mark, err := imgconv.Open(watermark)
		if err != nil {
			log.Fatal(err)
		}
		task.SetWatermark(mark, opacity)
		task.Watermark.SetRandom(random).SetOffset(image.Point{X: offsetX, Y: offsetY})
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
	}
	if !di.Mode().IsDir() {
		log.Fatal("Destination is not a directory.")
	}

	switch mode := si.Mode(); {
	case mode.IsDir():
		var message string
		var lastPath string
		var lastWidth int
		var images []string

		ticker := time.NewTicker(time.Second)
		done := make(chan bool)
		go func() {
			for {
				select {
				case <-done:
					ticker.Stop()
					return
				case <-ticker.C:
					m := message
					io.WriteString(
						os.Stderr,
						fmt.Sprintf("\r%s\r%s", strings.Repeat(" ", lastWidth), m),
					)
					lastWidth = len(m)
				}
			}
		}()

		filepath.WalkDir(src, func(path string, d fs.DirEntry, _ error) error {
			if ok, _ := regexp.MatchString(`^\.(jpe?g|png|gif|tiff?|bmp|pdf|webp)$`,
				strings.ToLower(filepath.Ext(d.Name()))); ok {
				images = append(images, path)
			}

			if d.IsDir() {
				lastPath = filepath.Dir(path)
			}
			message = fmt.Sprintf("Found images: %d, Scanning directory %s", len(images), lastPath)

			return nil
		})
		done <- true

		total := len(images)

		io.WriteString(os.Stderr, fmt.Sprintf("\r%s\r", strings.Repeat(" ", lastWidth)))
		log.Println("Total images:", total)

		pb := progressbar.New(total)
		pb.Start()
		workers.New(worker).Slice(images, func(_ int, i interface{}) {
			defer pb.Add(1)

			rel, _ := filepath.Rel(src, i.(string))
			output := task.ConvertExt(filepath.Join(dst, rel))
			if _, err := os.Stat(output); !os.IsNotExist(err) {
				log.Println("Skip", output)
				return
			}
			if err := os.MkdirAll(filepath.Dir(output), 0755); err != nil {
				log.Print(err)
				return
			}

			base, err := imgconv.Open(i.(string))
			if err != nil {
				log.Println(i, err)
				return
			}

			f, err := os.Create(output)
			if err != nil {
				log.Print(err)
				return
			}

			if err := task.Convert(f, base); err != nil {
				log.Println(i, err)
				defer os.Remove(output)
				return
			}
			defer f.Close()

			if debug {
				log.Printf("[Debug]Converted %s\n", i.(string))
			}
		})
		<-pb.Done

	case mode.IsRegular():
		output := task.ConvertExt(filepath.Join(dst, filepath.Base(src)))

		if _, err := os.Stat(output); !os.IsNotExist(err) {
			log.Fatal("Destination already exist.")
		}
		if err := os.MkdirAll(filepath.Dir(output), 0755); err != nil {
			log.Fatal(err)
		}

		base, err := imgconv.Open(src)
		if err != nil {
			log.Fatal(err)
		}

		f, err := os.Create(output)
		if err != nil {
			log.Fatal(err)
		}

		if err := task.Convert(f, base); err != nil {
			defer os.Remove(output)
			log.Fatal(err)
		}
		defer f.Close()

		log.Print("Done.")

	default:
		log.Fatal("Unknown source.")
	}

	fmt.Println("Press enter key to continue . . .")
	fmt.Scanln()
}
