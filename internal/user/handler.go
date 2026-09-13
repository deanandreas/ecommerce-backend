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
		httpx.Write(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	user, err := h.service.UserProfile(r.Context(), userID)
	if err != nil {
		if errors.Is(err, ErrNoRows) {
			httpx.Write(w, http.StatusNotFound, "not found", nil)
			return
		}
		slog.Error("filed to get user profile", "error", err)
		httpx.Write(w, http.StatusInternalServerError, "failed to fetch user profile", nil)
		return
	}

	httpx.Write(w, http.StatusOK, "profile fetched successfully", user)
}

func (h *Handler) UpdateUserProfile(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r)
	if err != nil {
		httpx.Write(w, http.StatusUnauthorized, "Unauthorized", nil)
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
			httpx.Write(w, http.StatusBadRequest, "no field provided to update", nil)
		case errors.Is(err, ErrInvalidPassword):
			httpx.Write(w, http.StatusBadRequest, "invalid password", nil)
		default:
			slog.Error("failed to update user profile", "error", err)
			httpx.Write(w, http.StatusInternalServerError, "failed to update user profile", nil)
		}
		return
	}

	httpx.Write(w, http.StatusOK, "profile updated successfully", user)
}

func (h *Handler) AddUserAddress(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r)
	if err != nil {
		httpx.Write(w, http.StatusUnauthorized, "Unauthorized", nil)
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
			httpx.Write(w, http.StatusBadRequest, "invaled address", nil)
		default:
			slog.Error("failed to add address", "error", err)
			httpx.Write(w, http.StatusInternalServerError, "failed to add address", nil)
		}
		return
	}

	httpx.Write(w, http.StatusCreated, "address added successfully", user)
}

func (h *Handler) UpdateDefaultAddress(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r)
	if err != nil {
		httpx.Write(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	id := r.PathValue("id")

	user, err := h.service.UpdateDefaultAddress(r.Context(), db.UpdateDefaultAddressParams{UserID: userID, ID: id})
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrInvalidData):
			httpx.Write(w, http.StatusBadRequest, "missing address id", nil)
		case errors.Is(err, ErrInvalidID):
			httpx.Write(w, http.StatusBadRequest, "invalid id", nil)
		default:
			slog.Error("failed to updated default address", "error", err)
			httpx.Write(w, http.StatusInternalServerError, "failed to change the default address", nil)
		}
		return
	}

	httpx.Write(w, http.StatusOK, "default address updated successfully", user)
}

func (h *Handler) DeleteAddress(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r)
	if err != nil {
		httpx.Write(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	id := r.PathValue("id")

	err = h.service.DeleteUserAddress(r.Context(), db.DeleteUserAddressParams{UserID: userID, ID: id})
	if err != nil {
		switch {
		case errors.Is(err, ErrNoRows):
			httpx.Write(w, http.StatusNotFound, "not found", nil)
		case errors.Is(err, ErrInvalidID):
			httpx.Write(w, http.StatusBadRequest, "invalid id", nil)
		default:
			slog.Error("failed to delete user address", "error", err)
			httpx.Write(w, http.StatusInternalServerError, "failed to delete the address", nil)
		}
		return
	}

	httpx.Write(w, http.StatusNoContent, "", nil)
}
