package database

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

func SaveUploads(r *http.Request, ext []string, folder, key string) ([]string, error) {
	// limite the size of request data (eg. 10GB-20GB)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		return nil, err
	}

	var fileURLs []string

	// open the requested file
	files := r.MultipartForm.File[key]
	if files == nil {
		return nil, ErrNoFile
	}

	for _, header := range files {
		file, err := header.Open()
		if err != nil {
			return nil, err
		}

		// to check the file type
		fileExt := strings.ToLower(filepath.Ext(header.Filename))
		isValid := false
		for i := range ext {
			if fileExt == ext[i] {
				isValid = true
			}
		}

		if !isValid {
			return nil, ErrInvalidType
		}

		re := regexp.MustCompile(`[^a-zA-Z.]`)
		s := re.ReplaceAllString(header.Filename, "")

		// TO MAKE UNQUE FILE
		fileName := fmt.Sprintf("%d-%s", time.Now().UnixNano(), s)
		uploadDir := filepath.Join("uploads", folder)
		// for local development only
		if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
			return nil, err
		}

		filepath := filepath.Join(uploadDir, fileName)
		dst, err := os.Create(filepath)
		if err != nil {
			return nil, err
		}

		if _, err := io.Copy(dst, file); err != nil {
			return nil, err
		}
		dst.Close()

		fileURL, _ := strings.CutPrefix(filepath, "uploads")

		fileURLs = append(fileURLs, fileURL)
		if err := file.Close(); err != nil {
			slog.Error("failed to close file", "error", err.Error())
		}
	}

	return fileURLs, nil
}

func DeleteUploads(path string) error {
	path = filepath.Join("uploads", path)
	if err := os.Remove(path); err != nil {
		return err
	}
	return nil
}

func RollBack(tx pgx.Tx, ctx context.Context) {
	if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
		slog.Error("failed to rollback", "error", err)
	}
}

func RollBackWithFile(tx pgx.Tx, ctx context.Context, files []string) {
	err := tx.Rollback(ctx)
	if err != nil && !errors.Is(err, pgx.ErrTxClosed) {
		slog.Error("failed to rollback", "error", err)
	}
	if !errors.Is(err, pgx.ErrTxClosed) {
		for _, path := range files {
			DeleteUploads(path)
		}
	}
}
