package server

import (
	"errors"
	"os"
	"strconv"

	"github.com/deanandreas/ecommerce-api/internal/storage"
	_ "github.com/joho/godotenv/autoload"
)

func GetENV() (int, string, error) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		return 0, "", errors.New("database URL does not exist in enviaroment variable")
	}

	portNum, _ := strconv.Atoi(port)

	return portNum, dbURL, nil
}

func GetMinIOConfig() storage.Config {
	return storage.Config{
		Endpoint:  os.Getenv("MINIO_ENDPOINT"),
		AccessKey: os.Getenv("MINIO_ACCESS_KEY"),
		SecretKey: os.Getenv("MINIO_SECRET_KEY"),
		Bucket:    os.Getenv("MINIO_BUCKET"),
		UseSSL:    os.Getenv("MINIO_USE_SSL") == "true",
	}
}
