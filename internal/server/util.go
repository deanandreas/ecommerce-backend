package server

import (
	"errors"
	"net/http"
	"os"
	"strconv"

	"github.com/deanandreas/ecommerce-api/internal/httpx"
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

func WriteJSON(w http.ResponseWriter, statusCode int, message string, data any) {
	httpx.Write(w, statusCode, message, data)
}

func ReadJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	return httpx.Read(w, r, dst)
}

func FromDataToJSON(r *http.Request, key string, dst any) error {
	return httpx.FromData(r, key, dst)
}
