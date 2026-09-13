package httpx

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
)

var (
	ErrMissingKey    = errors.New("key does not exist")
	ErrTypeUnmarshal *json.UnmarshalTypeError
)

type Response struct {
	Data    any    `json:"data"`
	Message string `json:"message"`
	Success bool   `json:"success"`
}

func Write(w http.ResponseWriter, statusCode int, message string, data any) {
	isOK := statusCode >= 200 && statusCode < 300

	resp := Response{
		Data:    data,
		Message: message,
		Success: isOK,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.Error("response error", "error", err.Error())
	}
}

func Read(w http.ResponseWriter, r *http.Request, dst any) bool {
	maxByte := int64(1048765)
	r.Body = http.MaxBytesReader(w, r.Body, maxByte)

	dec := json.NewDecoder(r.Body)

	if err := dec.Decode(&dst); err != nil {
		if errors.As(err, &ErrTypeUnmarshal) {
			Write(w, http.StatusBadRequest, "invalid value of json key", nil)
			return false
		}
		if errors.Is(err, io.EOF) {
			Write(w, http.StatusBadRequest, "body must contain json payload", nil)
			return false
		}
		slog.Error("failed to read json payload", "error", err)
		Write(w, http.StatusUnprocessableEntity, "invalid json payload", nil)
		return false
	}

	if err := dec.Decode(&struct{}{}); err != io.EOF {
		Write(w, http.StatusBadRequest, "too many json payload", nil)
		return false
	}

	return true
}

func FromData(r *http.Request, key string, dst any) error {
	if err := r.ParseMultipartForm(32 << 20); err != nil && err != http.ErrNotMultipart {
		return err
	}

	k := r.FormValue(key)
	if k == "" {
		return ErrMissingKey
	}

	if err := json.Unmarshal([]byte(k), dst); err != nil {
		return err
	}

	return nil
}
