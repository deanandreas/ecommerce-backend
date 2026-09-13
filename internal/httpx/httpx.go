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

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"status"`
}

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// Send writes data directly as the JSON response body with no envelope.
// A 204 status writes an empty body.
func Send(w http.ResponseWriter, status int, data any) {
	if status == http.StatusNoContent {
		w.WriteHeader(status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("response error", "error", err.Error())
	}
}

// Error writes a structured error envelope with a stable machine-readable code.
func Error(w http.ResponseWriter, status int, code, message string) {
	resp := ErrorResponse{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
			Status:  status,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
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
			Error(w, http.StatusBadRequest, "INVALID_JSON_VALUE", "invalid value of json key")
			return false
		}
		if errors.Is(err, io.EOF) {
			Error(w, http.StatusBadRequest, "BODY_REQUIRED", "body must contain json payload")
			return false
		}
		slog.Error("failed to read json payload", "error", err)
		Error(w, http.StatusUnprocessableEntity, "INVALID_JSON", "invalid json payload")
		return false
	}

	if err := dec.Decode(&struct{}{}); err != io.EOF {
		Error(w, http.StatusBadRequest, "TOO_MANY_JSON", "too many json payload")
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
