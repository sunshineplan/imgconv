package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"image"
	"image/color"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/sunshineplan/imgconv"
	"github.com/sunshineplan/utils/flags"
	"github.com/sunshineplan/utils/log"
	"github.com/sunshineplan/utils/progressbar"
	"github.com/sunshineplan/workers"
)

var (
	src             = flag.String("src", "", "")
	dst             = flag.String("dst", "output", "")
	test            = flag.Bool("test", false, "")
	force           = flag.Bool("force", false, "")
	pdf             = flag.Bool("pdf", false, "")
	whiteBackground = flag.Bool("white-background", false, "")
	gray            = flag.Bool("gray", false, "")
	quality         = flag.Int("quality", 75, "")
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
	quiet           = flag.Bool("q", false, "")
	debug           = flag.Bool("debug", false, "")

	format      imgconv.Format
	compression imgconv.TIFFCompression
)

func usage() {
	fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
	fmt.Println(`
  --src
		source file or directory
  --dst
		destination directory (default: output)
  --test
		test source file only, don't convert (default: false)
  --force
		force overwrite (default: false)
  --pdf
		convert pdf source (default: false)
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
			if err == flag.ErrHelp {
				return
			}
			log.Error("Panic", "error", err)
			if code == 0 {
				code = 1
			}
		}
		fmt.Println("Press enter key to exit . . .")
		fmt.Scanln()
		os.Exit(code)
	}()

	self, err := os.Executable()
	if err != nil {
		log.Error("Failed to get self path", "error", err)
		code = 1
		return
	}

	flag.CommandLine.Init(os.Args[0], flag.PanicOnError)
	flag.Usage = usage
	flag.TextVar(&format, "format", imgconv.JPEG, "")
	flag.TextVar(&compression, "compression", imgconv.TIFFDeflate, "")
	flags.SetConfigFile(filepath.Join(filepath.Dir(self), "config.ini"))
	flags.Parse()

	log.SetOutput(filepath.Join(filepath.Dir(self), fmt.Sprintf("convert%s.log", time.Now().Format("20060102150405"))), os.Stdout)
	if *debug {
		log.SetLevel(slog.LevelDebug)
	}

	srcInfo, err := os.Stat(*src)
	if err != nil {
		log.Error("Failed to get FileInfo of source", "source", *src, "error", err)
		code = 1
		return
	}

	if *test {
		switch {
		case srcInfo.Mode().IsDir():
			images := loadImages(*src, *pdf)
			total := len(images)
			log.Println("Total images:", total)
			pb := progressbar.New(total)
			pb.Start()
			workers.Workers(*worker).Run(context.Background(), workers.SliceJob(images, func(_ int, image string) {
				defer pb.Add(1)
				if _, err := open(image); err != nil {
					log.Error("Bad image", "image", image, "error", err)
				}
			}))
			pb.Done()
		case srcInfo.Mode().IsRegular():
			if _, err := open(*src); err != nil {
				log.Error("Bad image", "image", *src, "error", err)
			}
		default:
			log.Error("Unknown source mode", "mode", srcInfo.Mode())
			code = 1
		}
		return
	}

	task := imgconv.NewOptions()

	var opts []imgconv.EncodeOption
	if format == imgconv.JPEG || format == imgconv.PDF {
		opts = append(opts, imgconv.Quality(*quality))
	}
	if format == imgconv.TIFF {
		opts = append(opts, imgconv.TIFFCompressionType(compression))
	}
	if *whiteBackground {
		opts = append(opts, imgconv.BackgroundColor(color.White))
	}
	task.SetFormat(format, opts...)

	if *gray {
		task.SetGray(true)
	}
	if *watermark != "" {
		mark, err := imgconv.Open(*watermark)
		if err != nil {
			log.Error("Failed to open watermark", "watermark", *watermark, "error", err)
			code = 1
			return
		}
		task.SetWatermark(mark, *opacity)
		task.Watermark.SetRandom(*random).SetOffset(image.Point{X: *offsetX, Y: *offsetY})
	}
	if *width != 0 || *height != 0 || *percent != 0 {
		task.SetResize(*width, *height, *percent)
	}

	dstInfo, err := os.Stat(*dst)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			if err := os.MkdirAll(*dst, 0755); err != nil {
				log.Error("Failed to create directory for destination", "destination", *dst, "error", err)
				code = 1
				return
			}
			dstInfo, _ = os.Stat(*dst)
		} else {
			log.Error("Failed to get FileInfo of destination", "destination", *dst, "error", err)
			code = 1
			return
		}
	}

	switch {
	case srcInfo.Mode().IsDir():
		if !dstInfo.Mode().IsDir() {
			log.Error("Destination is not a directory", "destination", *dst)
			code = 1
			return
		}
		images := loadImages(*src, *pdf)
		total := len(images)
		log.Println("Total images:", total)
		pb := progressbar.New(total)
		if !*quiet {
			pb.Start()
		}
		workers.Workers(*worker).Run(context.Background(), workers.SliceJob(images, func(_ int, image string) {
			defer pb.Add(1)
			rel, err := filepath.Rel(*src, image)
			if err != nil {
				log.Error("Failed to get relative path", "source", *src, "image", image, "error", err)
				return
			}
			output := task.ConvertExt(filepath.Join(*dst, rel))
			if err := convert(task, image, output, *force); err != nil {
				if err == errSkip && !*quiet {
					log.Println("Skip", output)
				}
				return
			}
			log.Debug("Converted " + image)
		}))
		if !*quiet {
			pb.Done()
		}
	case srcInfo.Mode().IsRegular():
		output := *dst
		if dstInfo.Mode().IsDir() {
			output = task.ConvertExt(filepath.Join(output, srcInfo.Name()))
		}
		if err := convert(task, *src, output, *force); err != nil {
			if err == errSkip {
				log.Error("Destination already exist", "destination", output)
			}
			code = 1
			return
		}
	default:
		log.Error("Unknown source mode", "mode", srcInfo.Mode())
		code = 1
		return
	}
	log.Print("Done")
}
