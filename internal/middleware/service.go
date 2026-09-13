package middleware

import (
	"errors"
	"net/http"
)

type contextKey string

const userIDKey contextKey = "userID"

func GetUserID(r *http.Request) (string, error) {
	val, ok := r.Context().Value(userIDKey).(string)
	if !ok {
		return "", errors.New("failed to get the user id")
	}
	return val, nil
}
