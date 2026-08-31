package server

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type contextKey string

const (
	userIDKey   contextKey = "userID"
	userRoleKey contextKey = "role"
)

func (s *Server) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			WriteJSON(w, http.StatusUnauthorized, "missing authorization header", nil)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			WriteJSON(w, http.StatusUnauthorized, "Invalid token format structure", nil)
			return
		}
		tokenString := parts[1]

		userID, err := ValidateToken(tokenString)
		if err != nil {
			WriteJSON(w, http.StatusUnauthorized, "Invalid or expired token", nil)
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, userIDKey, userID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

// Override WriteHeader to capture the status code
func (rw *responseWriterWrapper) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (s *Server) LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		intercept := &responseWriterWrapper{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(intercept, r)
		duration := time.Since(startTime)
		slog.Info("completed request", "method", r.Method, "path", r.URL.Path, "status", intercept.statusCode, "duration", duration)
	})
}

func (s *Server) CORSMiddleware(nex http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Authorization", "Bearer <token>")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Contexnt-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		nex.ServeHTTP(w, r)
	})
}
