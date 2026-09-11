package server

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	_ "github.com/joho/godotenv/autoload"
)

func GetENV() (int, string, error) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbURL := os.Getenv("DBURL")
	if dbURL == "" {
		return 0, "", errors.New("database URL does not exist in enviaroment variable")
	}

	portNum, _ := strconv.Atoi(port)

	return portNum, dbURL, nil
}

func GetUserID(r *http.Request) (string, error) {
	val, ok := r.Context().Value(userIDKey).(string)
	if !ok {
		return "", errors.New("failed to get the user id")
	}
	return val, nil
}

type JSONResponse struct {
	Data    any    `json:"data"`
	Message string `json:"message"`
	Success bool   `json:"success"`
}

func WriteJSON(w http.ResponseWriter, statusCode int, message string, data any) {
	isOK := statusCode >= 200 && statusCode < 300

	resp := JSONResponse{
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

func ReadJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	maxByte := int64(1048765)
	r.Body = http.MaxBytesReader(w, r.Body, maxByte)

	dec := json.NewDecoder(r.Body)

	if err := dec.Decode(&dst); err != nil {
		if errors.As(err, &ErrTypeUnmarshal) {
			WriteJSON(w, http.StatusBadRequest, "invalid value of json key", nil)
			return false
		}
		if errors.Is(err, io.EOF) {
			WriteJSON(w, http.StatusBadRequest, "body must contain json payload", nil)
			return false
		}
		slog.Error("failed to read json payload", "error", err)
		WriteJSON(w, http.StatusUnprocessableEntity, "invalid json payload", nil)
		return false
	}

	if err := dec.Decode(&struct{}{}); err != io.EOF {
		WriteJSON(w, http.StatusBadRequest, "too many json payload", nil)
		return false
	}

	return true
}

func FromDataToJSON(r *http.Request, key string, dst any) error {
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
