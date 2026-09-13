package auth

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/deanandreas/ecommerce-api/internal/database"
	"github.com/deanandreas/ecommerce-api/internal/httpx"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var user database.UserData
	if !httpx.Read(w, r, &user) {
		return
	}

	res, err := h.service.Register(r.Context(), user, user.Password)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidData):
			httpx.Write(w, http.StatusBadRequest, "invalid form of data", nil)
		case errors.Is(err, ErrEmailExists):
			httpx.Write(w, http.StatusConflict, "email already in use", nil)
		default:
			slog.Error("failed to register the user", "error", err)
			httpx.Write(w, http.StatusInternalServerError, "failed to register the user", nil)
		}
		return
	}

	httpx.Write(w, http.StatusCreated, "user created successfully", res)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if !httpx.Read(w, r, &req) {
		return
	}

	res, err := h.service.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidData):
			httpx.Write(w, http.StatusBadRequest, "email and password are required", nil)
		case errors.Is(err, ErrInvalidCredentials):
			httpx.Write(w, http.StatusUnauthorized, "invalid email or password", nil)
		default:
			slog.Error("failed to login", "error", err)
			httpx.Write(w, http.StatusInternalServerError, "failed to login", nil)
		}
		return
	}

	httpx.Write(w, http.StatusOK, "user login successfully", res)
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if !httpx.Read(w, r, &req) {
		return
	}

	res, err := h.service.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidData):
			httpx.Write(w, http.StatusBadRequest, "missing refresh token", nil)
		case errors.Is(err, ErrRefreshTokenNotFound):
			httpx.Write(w, http.StatusUnauthorized, "refresh token does not exist", nil)
		case errors.Is(err, ErrRefreshTokenExpired):
			httpx.Write(w, http.StatusLocked, "expiared cookie", nil)
		default:
			slog.Error("failed to refresh token", "error", err)
			httpx.Write(w, http.StatusInternalServerError, "failed to generated token", nil)
		}
		return
	}

	httpx.Write(w, http.StatusOK, "token generated successfully", res)
}
