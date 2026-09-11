package server

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	_ "github.com/joho/godotenv/autoload"
	"golang.org/x/crypto/bcrypt"
)

func getJWTKey() ([]byte, error) {
	key := os.Getenv("JWT_KEY")
	if key == "" {
		return nil, errors.New("JWT_KEY enviaroment viriable is not set")
	}
	return []byte(key), nil
}

func HashPassword(password string) (string, error) {
	byte, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	hashPassword := string(byte)

	return hashPassword, nil
}

func ValidatePassword(hashPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashPassword), []byte(password))
	return err == nil
}

func GenerateJWT(userID string) (string, error) {
	key, err := getJWTKey()
	if err != nil {
		return "", err
	}

	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(5 * 24 * time.Hour).Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	stringToken, err := token.SignedString(key)
	if err != nil {
		return "", fmt.Errorf("failed to create JWT: %w", err)
	}

	return stringToken, nil
}

func ValidateToken(tokenString string) (string, error) {
	key, err := getJWTKey()
	if err != nil {
		return "", err
	}

	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return key, nil
	})

	if err != nil || !token.Valid {
		return "", errors.New("invalid or expired token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("failed to parce token claims")
	}

	userID, ok := claims["sub"].(string)
	if !ok || userID == "" {
		return "", errors.New("token is missing subject identifier")
	}

	return userID, nil
}

func HashRefreshToken(val string) string {
	hasher := sha256.New()

	hasher.Write([]byte(val))

	hasherByte := hasher.Sum(nil)
	return hex.EncodeToString(hasherByte)
}
