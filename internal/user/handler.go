package user

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/deanandreas/ecommerce-api/internal/auth"
	db "github.com/deanandreas/ecommerce-api/internal/database/sqlc"
	"github.com/deanandreas/ecommerce-api/internal/httpx"
	"github.com/deanandreas/ecommerce-api/internal/middleware"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r)
	if err != nil {
		httpx.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized")
		return
	}

	user, err := h.service.UserProfile(r.Context(), userID)
	if err != nil {
		if errors.Is(err, ErrNoRows) {
			httpx.Error(w, http.StatusNotFound, "PROFILE_NOT_FOUND", "not found")
			return
		}
		slog.Error("filed to get user profile", "error", err)
		httpx.Error(w, http.StatusInternalServerError, "INTERNAL", "failed to fetch user profile")
		return
	}

	httpx.Send(w, http.StatusOK, user)
}

func (h *Handler) UpdateUserProfile(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r)
	if err != nil {
		httpx.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized")
		return
	}

	var req UpdateProfile
	if !httpx.Read(w, r, &req) {
		return
	}

	req.ID = userID
	ctx := r.Context()

	user, err := h.service.UpdateProfile(ctx, req)
	if err != nil {
		switch {
		case errors.Is(err, ErrNoFiled):
			httpx.Error(w, http.StatusBadRequest, "NO_FIELDS_TO_UPDATE", "no field provided to update")
		case errors.Is(err, ErrInvalidPassword):
			httpx.Error(w, http.StatusBadRequest, "INVALID_PASSWORD", "invalid password")
		default:
			slog.Error("failed to update user profile", "error", err)
			httpx.Error(w, http.StatusInternalServerError, "INTERNAL", "failed to update user profile")
		}
		return
	}

	httpx.Send(w, http.StatusOK, user)
}

func (h *Handler) AddUserAddress(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r)
	if err != nil {
		httpx.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized")
		return
	}

	var req db.InsertAddressParams
	if !httpx.Read(w, r, &req) {
		return
	}
	req.UserID = userID

	user, err := h.service.AddAddress(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrInvalidData):
			httpx.Error(w, http.StatusBadRequest, "INVALID_ADDRESS", "invaled address")
		default:
			slog.Error("failed to add address", "error", err)
			httpx.Error(w, http.StatusInternalServerError, "INTERNAL", "failed to add address")
		}
		return
	}

	httpx.Send(w, http.StatusCreated, user)
}

func (h *Handler) UpdateDefaultAddress(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r)
	if err != nil {
		httpx.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized")
		return
	}

	id := r.PathValue("id")

	user, err := h.service.UpdateDefaultAddress(r.Context(), db.UpdateDefaultAddressParams{UserID: userID, ID: id})
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrInvalidData):
			httpx.Error(w, http.StatusBadRequest, "ADDRESS_ID_REQUIRED", "missing address id")
		case errors.Is(err, ErrInvalidID):
			httpx.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid id")
		default:
			slog.Error("failed to updated default address", "error", err)
			httpx.Error(w, http.StatusInternalServerError, "INTERNAL", "failed to change the default address")
		}
		return
	}

	httpx.Send(w, http.StatusOK, user)
}

func (h *Handler) DeleteAddress(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r)
	if err != nil {
		httpx.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized")
		return
	}

	id := r.PathValue("id")

	err = h.service.DeleteUserAddress(r.Context(), db.DeleteUserAddressParams{UserID: userID, ID: id})
	if err != nil {
		switch {
		case errors.Is(err, ErrNoRows):
			httpx.Error(w, http.StatusNotFound, "ADDRESS_NOT_FOUND", "not found")
		case errors.Is(err, ErrInvalidID):
			httpx.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid id")
		default:
			slog.Error("failed to delete user address", "error", err)
			httpx.Error(w, http.StatusInternalServerError, "INTERNAL", "failed to delete the address")
		}
		return
	}

	httpx.Send(w, http.StatusNoContent, nil)
}
