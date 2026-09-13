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
		httpx.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
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
			httpx.Error(w, http.StatusBadRequest, "INVALID_REVIEW_INPUT", "product id and rating are required")
			return
		}
		slog.Error("failed to insert review", "error", err)
		httpx.Error(w, http.StatusInternalServerError, "INTERNAL", "failed to save the review")
		return
	}

	httpx.Send(w, http.StatusCreated, review)
}

func (h *Handler) GetUserReview(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r)
	if err != nil {
		httpx.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}
	id := r.PathValue("id")
	if id == "" {
		httpx.Error(w, http.StatusBadRequest, "PRODUCT_ID_REQUIRED", "product id is reqeured")
		return
	}

	review, err := h.service.GetUserReview(r.Context(), db.GetUserReviewParams{UserID: userID, ProductID: id})
	if err != nil {
		slog.Error("failed to get user review", "error", err)
		httpx.Error(w, http.StatusInternalServerError, "INTERNAL", "failed to get user Review")
		return
	}

	httpx.Send(w, http.StatusOK, review)
}

func (h *Handler) GetProductReviews(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		httpx.Error(w, http.StatusBadRequest, "PRODUCT_ID_REQUIRED", "product id is reqeured")
		return
	}

	reviews, err := h.service.GetProductReviews(r.Context(), id)
	if err != nil {
		slog.Error("failed to get user review", "error", err)
		httpx.Error(w, http.StatusInternalServerError, "INTERNAL", "failed to get user Review")
		return
	}

	httpx.Send(w, http.StatusOK, reviews)
}
