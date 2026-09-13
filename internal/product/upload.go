package product

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/deanandreas/ecommerce-api/internal/storage"
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

func saveImages(ctx context.Context, store storage.ImageStore, r *http.Request) ([]string, error) {
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

		fileExt := strings.TrimPrefix(strings.ToLower(filepath.Ext(header.Filename)), ".")
		isValid := false
		for _, ext := range Ext {
			if strings.TrimPrefix(ext, ".") == fileExt {
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
		case "png":
			format = imaging.PNG
		case "gif":
			format = imaging.GIF
		default:
			format = imaging.JPEG
		}

		var buf bytes.Buffer
		if err := imaging.Encode(&buf, dstImage, format); err != nil {
			return nil, err
		}

		fileName := fmt.Sprintf("%d-%s", time.Now().UnixNano(), s)
		key := imageFolder + "/" + fileName

		if err := store.Put(ctx, key, buf.Bytes()); err != nil {
			return nil, err
		}

		fileURLs = append(fileURLs, "/"+key)
		if err := file.Close(); err != nil {
			slog.Error("failed to close file", "error", err.Error())
		}
	}

	return fileURLs, nil
}

func removeFile(ctx context.Context, store storage.ImageStore, path string) {
	key := strings.TrimPrefix(path, "/")
	if err := store.Remove(ctx, key); err != nil {
		slog.Error("failed to remove image", "error", err)
	}
}

func deleteFiles(ctx context.Context, store storage.ImageStore, paths []string) {
	for _, path := range paths {
		removeFile(ctx, store, path)
	}
}
