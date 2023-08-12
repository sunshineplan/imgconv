package main

import (
	"errors"
	"flag"
	"fmt"
	"image"
	"image/color"
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
	"github.com/sunshineplan/utils/flags"
	"github.com/sunshineplan/utils/progressbar"
	"github.com/sunshineplan/utils/workers"
)

var (
	supported = regexp.MustCompile(`(?i)\.(jpe?g|png|gif|tiff?|bmp|pdf|webp)$`)
	tiffImage = regexp.MustCompile(`(?i)\.tiff?$`)
)

var (
	src             = flag.String("src", "", "")
	dst             = flag.String("dst", "output", "")
	force           = flag.Bool("force", false, "")
	format          = flag.String("format", "jpg", "")
	whiteBackground = flag.Bool("white-background", false, "")
	gray            = flag.Bool("gray", false, "")
	quality         = flag.Int("quality", 75, "")
	compression     = flag.String("compression", "deflate", "")
	autoOrientation = flag.Bool("auto-orientation", false, "")
	watermark       = flag.String("watermark", "", "")
	opacity         = flag.Uint("opacity", 128, "")
	random          = flag.Bool("random", false, "")
	offsetX         = flag.Int("x", 0, "")
	offsetY         = flag.Int("y", 0, "")
	width           = flag.Int("width", 0, "")
	height          = flag.Int("height", 0, "")
	percent         = flag.Float64("percent", 0, "")
	worker          = flag.Int("worker", 5, "")
	debug           = flag.Bool("debug", false, "")
)

func usage() {
	fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
	fmt.Println(`
  --src
		source file or directory
  --dst
		destination directory (default: output)
  --force
		force overwrite (default: false)
  --format
		output format (jpg, jpeg, png, gif, tif, tiff, bmp and pdf are supported, default: jpg)
  --white-background
		use white color for transparent background (default: false)
  --gray
		convert to grayscale (default: false)
  --quality
		set jpeg or pdf quality (range 1-100, default: 75)
  --compression
		set tiff compression type (none, deflate, default: deflate)
  --auto-orientation
		auto orientation (default: false)
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
	var code int
	defer func() {
		if err := recover(); err != nil {
			log.Print(err)
		}
		fmt.Println("Press enter key to exit . . .")
		fmt.Scanln()
		os.Exit(code)
	}()

	self, err := os.Executable()
	if err != nil {
		log.Println("Failed to get self path", err)
		code = 1
		return
	}

	flag.Usage = usage
	flags.SetConfigFile(filepath.Join(filepath.Dir(self), "config.ini"))
	flags.Parse()

	f, err := os.OpenFile(
		filepath.Join(filepath.Dir(self), fmt.Sprintf("convert%s.log", time.Now().Format("20060102150405"))),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println("Failed to open log file:", err)
		code = 1
		return
	}
	defer f.Close()

	log.SetOutput(io.MultiWriter(f, os.Stdout))

	task := imgconv.NewOptions()

	var ct imgconv.TIFFCompression
	switch strings.ToLower(*compression) {
	case "none":
		ct = imgconv.TIFFUncompressed
	case "deflate":
		ct = imgconv.TIFFDeflate
	default:
		log.Println("Unknown tiff compression type:", ct)
		code = 1
		return
	}

	format, err := imgconv.FormatFromExtension(*format)
	if err != nil {
		log.Print(err)
		code = 1
		return
	}
	if *whiteBackground {
		task.SetFormat(format, imgconv.Quality(*quality), imgconv.TIFFCompressionType(ct), imgconv.BackgroundColor(color.White))
	} else {
		task.SetFormat(format, imgconv.Quality(*quality), imgconv.TIFFCompressionType(ct))
	}

	if *gray {
		task.SetGray(true)
	}

	if *watermark != "" {
		mark, err := imgconv.Open(*watermark)
		if err != nil {
			log.Print(err)
			code = 1
			return
		}
		task.SetWatermark(mark, *opacity)
		task.Watermark.SetRandom(*random).SetOffset(image.Point{X: *offsetX, Y: *offsetY})
	}
	if *width != 0 || *height != 0 || *percent != 0 {
		task.SetResize(*width, *height, *percent)
	}

	srcInfo, err := os.Stat(*src)
	if err != nil {
		log.Print(err)
		code = 1
		return
	}

	dstInfo, err := os.Stat(*dst)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			if err := os.MkdirAll(*dst, 0755); err != nil {
				log.Print(err)
				code = 1
				return
			}
			dstInfo, _ = os.Stat(*dst)
		} else {
			log.Print(err)
			code = 1
			return
		}
	}
	if !dstInfo.Mode().IsDir() {
		log.Print("Destination is not a directory.")
		code = 1
		return
	}

	switch mode := srcInfo.Mode(); {
	case mode.IsDir():
		var message string
		var lastPath string
		var lastWidth int
		var images []string

		ticker := time.NewTicker(time.Second)
		done := make(chan struct{})
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

		filepath.WalkDir(*src, func(path string, d fs.DirEntry, _ error) error {
			if supported.MatchString(d.Name()) {
				images = append(images, path)
			}

			if d.IsDir() {
				lastPath = filepath.Dir(path)
			}
			message = fmt.Sprintf("Found images: %d, Scanning directory %s", len(images), lastPath)

			return nil
		})
		close(done)

		total := len(images)

		io.WriteString(os.Stderr, fmt.Sprintf("\r%s\r", strings.Repeat(" ", lastWidth)))
		log.Println("Total images:", total)

		pb := progressbar.New(total)
		pb.Start()
		workers.RunSlice(*worker, images, func(_ int, image string) {
			defer func() {
				if err := recover(); err != nil {
					log.Println(image, err)
				}
				pb.Add(1)
			}()

			rel, err := filepath.Rel(*src, image)
			if err != nil {
				log.Println(image, err)
				return
			}
			output := task.ConvertExt(filepath.Join(*dst, rel))
			path := filepath.Dir(output)

			if _, err := os.Stat(output); !errors.Is(err, fs.ErrNotExist) && !*force {
				log.Println("Skip", output)
				return
			}
			if err := os.MkdirAll(path, 0755); err != nil {
				log.Print(err)
				return
			}

			img, err := open(image)
			if err != nil {
				log.Println(image, err)
				return
			}

			f, err := os.CreateTemp(path, "*.tmp")
			if err != nil {
				log.Print(err)
				return
			}

			if err := task.Convert(f, img); err != nil {
				log.Println(image, err)
				return
			}
			f.Close()

			os.Rename(f.Name(), output)

			if *debug {
				log.Printf("[Debug]Converted %s\n", image)
			}
		})
		pb.Done()

	case mode.IsRegular():
		output := task.ConvertExt(filepath.Join(*dst, filepath.Base(*src)))
		path := filepath.Dir(output)

		if _, err := os.Stat(output); !errors.Is(err, fs.ErrNotExist) && !*force {
			log.Print("Destination already exist.")
			code = 1
			return
		}
		if err := os.MkdirAll(path, 0755); err != nil {
			log.Print(err)
			code = 1
			return
		}

		base, err := open(*src)
		if err != nil {
			log.Print(err)
			code = 1
			return
		}

		f, err := os.CreateTemp(path, "*.tmp")
		if err != nil {
			log.Print(err)
			code = 1
			return
		}

		if err := task.Convert(f, base); err != nil {
			log.Print(err)
			code = 1
			return
		}
		f.Close()

		os.Rename(f.Name(), output)

	default:
		log.Print("Unknown source.")
		code = 1
	}
	log.Print("Done.")
}

func open(file string) (image.Image, error) {
	img, err := imgconv.Open(file, imgconv.AutoOrientation(*autoOrientation))
	if err != nil {
		if tiffImage.MatchString(file) {
			f, err := os.Open(file)
			if err != nil {
				return nil, err
			}
			defer f.Close()
			return tiff.Decode(f)
		}
	}
	return img, nil
}
