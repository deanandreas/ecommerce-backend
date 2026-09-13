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
			httpx.Error(w, http.StatusBadRequest, "INVALID_FORM_DATA", "invalid form of data")
		case errors.Is(err, ErrEmailExists):
			httpx.Error(w, http.StatusConflict, "EMAIL_IN_USE", "email already in use")
		default:
			slog.Error("failed to register the user", "error", err)
			httpx.Error(w, http.StatusInternalServerError, "INTERNAL", "failed to register the user")
		}
		return
	}

	httpx.Send(w, http.StatusCreated, res)
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
			httpx.Error(w, http.StatusBadRequest, "AUTH_FIELDS_REQUIRED", "email and password are required")
		case errors.Is(err, ErrInvalidCredentials):
			httpx.Error(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "invalid email or password")
		default:
			slog.Error("failed to login", "error", err)
			httpx.Error(w, http.StatusInternalServerError, "INTERNAL", "failed to login")
		}
		return
	}

	httpx.Send(w, http.StatusOK, res)
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
			httpx.Error(w, http.StatusBadRequest, "MISSING_REFRESH_TOKEN", "missing refresh token")
		case errors.Is(err, ErrRefreshTokenNotFound):
			httpx.Error(w, http.StatusUnauthorized, "INVALID_REFRESH_TOKEN", "refresh token does not exist")
		case errors.Is(err, ErrRefreshTokenExpired):
			httpx.Error(w, http.StatusLocked, "REFRESH_TOKEN_EXPIRED", "expiared cookie")
		default:
			slog.Error("failed to refresh token", "error", err)
			httpx.Error(w, http.StatusInternalServerError, "INTERNAL", "failed to generated token")
		}
		return
	}

	httpx.Send(w, http.StatusOK, res)
}
