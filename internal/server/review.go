package server

import (
	"log/slog"
	"net/http"

	db "github.com/deanandreas/ecommerce-api/internal/database/sqlc"
)

func (s *Server) CreateReview(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		WriteJSON(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	var req db.InsertReviewParams
	if ok := ReadJSON(w, r, &req); !ok {
		return
	}

	if req.ProductID == "" || req.Rating == 0 {
		WriteJSON(w, http.StatusBadRequest, "product id and rating are required", nil)
		return
	}
	req.UserID = userID

	ctx := r.Context()
	review, err := s.db.InsertReviewTx(ctx, req)
	if err != nil {
		slog.Error("failed to insert review", "error", err)
		WriteJSON(w, http.StatusInternalServerError, "failed to save the review", nil)
		return
	}

	WriteJSON(w, http.StatusCreated, "product reviewd successfully", review)
}

func (s *Server) GetUserReview(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		WriteJSON(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}
	id := r.PathValue("id")
	if id == "" {
		WriteJSON(w, http.StatusBadRequest, "product id is reqeured", nil)
		return
	}

	ctx := r.Context()
	review, err := s.db.GetUserReview(ctx, db.GetUserReviewParams{UserID: userID, ProductID: id})
	if err != nil {
		slog.Error("failed to get user review", "error", err)
		WriteJSON(w, http.StatusInternalServerError, "failed to get user Review", nil)
		return
	}

	WriteJSON(w, http.StatusOK, "review fetched successfully", review)
}

func (s *Server) GetProductReviews(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		WriteJSON(w, http.StatusBadRequest, "product id is reqeured", nil)
		return
	}

	ctx := r.Context()
	reviews, err := s.db.GetProductReviews(ctx, id)
	if err != nil {
		slog.Error("failed to get user review", "error", err)
		WriteJSON(w, http.StatusInternalServerError, "failed to get user Review", nil)
		return
	}

	WriteJSON(w, http.StatusOK, "review fetched successfully", reviews)
}
