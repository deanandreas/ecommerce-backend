package storage

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/minio/minio-go/v7"
)

// ImageHandler streams a product image object from the store.
//
// Mount it behind http.StripPrefix("/api/v1/products/image/", ...) so the
// request path is the file name, e.g. "123.png". The object key is built as
// "products/image/<file>" to match how product images are stored.
func ImageHandler(store ImageStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rc, size, err := store.Get(r.Context(), "products/image/"+r.URL.Path)
		if err != nil {
			if minio.ToErrorResponse(err).Code == "NoSuchKey" {
				http.NotFound(w, r)
				return
			}
			slog.Error("failed to get image", "error", err)
			http.Error(w, "failed to get image", http.StatusInternalServerError)
			return
		}
		defer rc.Close()

		w.Header().Set("Content-Type", contentTypeFor(r.URL.Path))
		w.Header().Set("Content-Length", fmt.Sprintf("%d", size))
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		w.WriteHeader(http.StatusOK)

		if _, err := io.Copy(w, rc); err != nil {
			slog.Error("failed to stream image", "error", err)
		}
	})
}
