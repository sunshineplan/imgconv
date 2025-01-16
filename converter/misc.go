package main

import (
	"errors"
	"fmt"
	"image"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/sunshineplan/imgconv"
	"github.com/sunshineplan/tiff"
	"github.com/sunshineplan/utils/log"
)

var (
	supported = regexp.MustCompile(`(?i)\.(jpe?g|png|gif|tiff?|bmp|webp)$`)
	pdfImage  = regexp.MustCompile(`(?i)\.pdf$`)
	tiffImage = regexp.MustCompile(`(?i)\.tiff?$`)
)

func open(file string) (image.Image, error) {
	img, err := imgconv.Open(file, imgconv.AutoOrientation(*autoOrientation))
	if err != nil && tiffImage.MatchString(file) {
		f, err := os.Open(file)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		return tiff.Decode(f)
	}
	return img, err
}

func loadImages(root string, pdf bool) (imgs []string) {
	var message string
	var width int
	done := make(chan struct{})
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				m := message
				if !*quiet {
					fmt.Fprintf(os.Stdout, "\r%s\r%s", strings.Repeat(" ", width), m)
				}
				width = len(m)
			}
		}
	}()
	var dir string
	filepath.WalkDir(root, func(path string, d fs.DirEntry, _ error) error {
		if supported.MatchString(d.Name()) || (pdf && pdfImage.MatchString(d.Name())) {
			imgs = append(imgs, path)
		}
		if d.IsDir() {
			dir = filepath.Dir(path)
		}
		message = fmt.Sprintf("Found images: %d, Scanning directory %s", len(imgs), dir)
		return nil
	})
	close(done)
	if !*quiet {
		fmt.Fprintf(os.Stdout, "\r%s\r", strings.Repeat(" ", width))
	}
	return
}

var errSkip = errors.New("skip")

func convert(task *imgconv.Options, image, output string, force bool) (err error) {
	if _, err = os.Stat(output); err == nil {
		if !force {
			return errSkip
		}
	} else if !errors.Is(err, fs.ErrNotExist) {
		log.Error("Failed to get FileInfo", "name", output, "error", err)
		return
	}
	path := filepath.Dir(output)
	if err = os.MkdirAll(path, 0755); err != nil {
		log.Error("Failed to create directory", "path", path, "error", err)
		return
	}
	img, err := open(image)
	if err != nil {
		log.Error("Failed to open image", "image", image, "error", err)
		return
	}
	f, err := os.CreateTemp(path, "*.tmp")
	if err != nil {
		log.Error("Failed to create temporary file", "path", path, "error", err)
		return
	}
	if err = task.Convert(f, img); err != nil {
		log.Error("Failed to convert image", "image", image, "error", err)
		return
	}
	f.Close()
	if err = os.Rename(f.Name(), output); err != nil {
		log.Error("Failed to move file", "from", f.Name(), "to", output, "error", err)
	}
	return
}
