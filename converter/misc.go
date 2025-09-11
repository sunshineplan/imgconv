package main

import (
	"errors"
	"fmt"
	"image"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sunshineplan/imgconv"
	"github.com/sunshineplan/tiff"
	"github.com/sunshineplan/utils/log"
)

var (
	supported = []string{".jpg", ".jpeg", ".png", ".gif", ".tif", ".tiff", ".bmp", ".webp"}
	pdfImage  = []string{".pdf"}
	tiffImage = []string{".tif", ".tiff"}
)

func matchFile(exts []string, file string) bool {
	file = strings.ToLower(file)
	for _, i := range exts {
		if strings.HasSuffix(file, i) {
			return true
		}
	}
	return false
}

func open(file string) (image.Image, error) {
	img, err := imgconv.Open(file, imgconv.AutoOrientation(*autoOrientation))
	if err != nil && matchFile(tiffImage, file) {
		f, err := os.Open(file)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		return tiff.Decode(f)
	}
	return img, err
}

func size(file string) (n int64) {
	info, err := os.Stat(file)
	if err == nil {
		n = info.Size()
	}
	return
}

func shorten(path string) string {
	if runes := []rune(path); len(runes) > 50 {
		return string(runes[:25]) + " ... " + string(runes[len(runes)-25:])
	}
	return path
}

func loadImages(root string, pdf bool) (imgs []string, size int64) {
	c := make(chan walkerResult)
	done := make(chan struct{})
	var dir string
	var message string
	go func() {
		for {
			res, ok := <-c
			if !ok {
				close(done)
				return
			}
			if res.isDir {
				dir = res.path
			} else {
				imgs = append(imgs, res.path)
				size += res.size
			}
			message = fmt.Sprintf("Found images: %d, Scanning directory %s", len(imgs), shorten(dir))
		}
	}()
	if !*quiet {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		go func() {
			var width int
			for {
				select {
				case <-done:
					fmt.Fprintf(os.Stdout, "\r%s\r", strings.Repeat(" ", width))
					return
				case <-ticker.C:
					m := message
					fmt.Fprintf(os.Stdout, "\r%s\r%s", strings.Repeat(" ", width), m)
					width = len(m)
				}
			}
		}()
	}
	walkDir(root, pdf, c)
	return
}

type walkerResult struct {
	path  string
	size  int64
	isDir bool
}

func walkDir(root string, pdf bool, c chan<- walkerResult) {
	defer close(c)
	filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Print(err)
			return nil
		}
		if d.IsDir() {
			c <- walkerResult{path: path, isDir: true}
		} else if name := d.Name(); matchFile(supported, name) || (pdf && matchFile(pdfImage, name)) {
			info, err := d.Info()
			if err != nil {
				log.Error("Failed to get FileInfo", "name", path, "error", err)
				return nil
			}
			c <- walkerResult{path: path, size: info.Size()}
		}
		return nil
	})
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
	err = task.Convert(f, img)
	f.Close()
	if err != nil {
		log.Error("Failed to convert image", "image", image, "error", err)
		return
	}
	if err = os.Rename(f.Name(), output); err != nil {
		log.Error("Failed to move file", "from", f.Name(), "to", output, "error", err)
	}
	return
}
