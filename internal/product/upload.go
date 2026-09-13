package product

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/disintegration/imaging"
)

var (
	ErrInvalidType = errors.New("invalid file type")
	ErrNoFile      = errors.New("no file found")
)

var Ext = []string{".png", ".gif", ".jpeg", ".jpg"}

const (
	imageFolder = "products/image"
	imageKey    = "image"
)

func saveImages(r *http.Request) ([]string, error) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		return nil, err
	}

	var fileURLs []string

	files := r.MultipartForm.File[imageKey]
	if files == nil {
		return nil, ErrNoFile
	}

	for _, header := range files {
		file, err := header.Open()
		if err != nil {
			return nil, err
		}

		fileExt := strings.ToLower(filepath.Ext(header.Filename))
		isValid := false
		for _, ext := range Ext {
			if fileExt == ext {
				isValid = true
			}
		}
		if !isValid {
			return nil, ErrInvalidType
		}

		srcImage, err := imaging.Decode(file)
		if err != nil {
			return nil, err
		}

		dstImage := imaging.Resize(srcImage, 300, 0, imaging.Lanczos)
		re := regexp.MustCompile(`[^a-zA-Z.]`)
		s := re.ReplaceAllString(header.Filename, "")

		var format imaging.Format
		switch fileExt {
		case ".png":
			format = imaging.PNG
		case ".gif":
			format = imaging.GIF
		default:
			format = imaging.JPEG
		}

		fileName := fmt.Sprintf("%d-%s", time.Now().UnixNano(), s)
		uploadDir := filepath.Join("uploads", imageFolder)
		if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
			return nil, err
		}

		path := filepath.Join(uploadDir, fileName)
		dst, err := os.Create(path)
		if err != nil {
			return nil, err
		}

		if err := imaging.Encode(dst, dstImage, format); err != nil {
			return nil, err
		}
		dst.Close()

		fileURL, _ := strings.CutPrefix(path, "uploads")
		fileURLs = append(fileURLs, fileURL)
		if err := file.Close(); err != nil {
			slog.Error("failed to close file", "error", err.Error())
		}
	}

	return fileURLs, nil
}

func deleteUpload(path string) error {
	path = filepath.Join("uploads", path)
	if err := os.Remove(path); err != nil {
		return err
	}
	return nil
}

func deleteFiles(paths []string) {
	for _, path := range paths {
		if err := deleteUpload(path); err != nil {
			slog.Error("failed to delete file", "error", err)
		}
	}
}
