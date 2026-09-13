package review

import (
	"errors"
	"log/slog"
	"net/http"

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

func (h *Handler) CreateReview(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r)
	if err != nil {
		httpx.Write(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	var req db.InsertReviewParams
	if !httpx.Read(w, r, &req) {
		return
	}
	req.UserID = userID

	review, err := h.service.CreateReview(r.Context(), req)
	if err != nil {
		if errors.Is(err, ErrInvalidInput) {
			httpx.Write(w, http.StatusBadRequest, "product id and rating are required", nil)
			return
		}
		slog.Error("failed to insert review", "error", err)
		httpx.Write(w, http.StatusInternalServerError, "failed to save the review", nil)
		return
	}

	httpx.Write(w, http.StatusCreated, "product reviewd successfully", review)
}

func (h *Handler) GetUserReview(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r)
	if err != nil {
		httpx.Write(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}
	id := r.PathValue("id")
	if id == "" {
		httpx.Write(w, http.StatusBadRequest, "product id is reqeured", nil)
		return
	}

	review, err := h.service.GetUserReview(r.Context(), db.GetUserReviewParams{UserID: userID, ProductID: id})
	if err != nil {
		slog.Error("failed to get user review", "error", err)
		httpx.Write(w, http.StatusInternalServerError, "failed to get user Review", nil)
		return
	}

	httpx.Write(w, http.StatusOK, "review fetched successfully", review)
}

func (h *Handler) GetProductReviews(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		httpx.Write(w, http.StatusBadRequest, "product id is reqeured", nil)
		return
	}

	reviews, err := h.service.GetProductReviews(r.Context(), id)
	if err != nil {
		slog.Error("failed to get user review", "error", err)
		httpx.Write(w, http.StatusInternalServerError, "failed to get user Review", nil)
		return
	}

	httpx.Write(w, http.StatusOK, "review fetched successfully", reviews)
}
