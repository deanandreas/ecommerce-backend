package payment

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/deanandreas/ecommerce-api/internal/database"
	"github.com/deanandreas/ecommerce-api/internal/httpx"
	"github.com/deanandreas/ecommerce-api/internal/middleware"
	"github.com/jackc/pgx/v5"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) InitiatePayment(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r)
	if err != nil {
		httpx.Write(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	orderID := r.PathValue("id")
	if orderID == "" {
		httpx.Write(w, http.StatusBadRequest, "order id required", nil)
		return
	}

	payment, err := h.service.InitiatePayment(r.Context(), orderID, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httpx.Write(w, http.StatusBadRequest, "no pending order found", nil)
			return
		}
		if errors.Is(err, ErrOrderNotPending) {
			httpx.Write(w, http.StatusBadRequest, "order is not pending", nil)
			return
		}
		slog.Error("failed to initiate payment", "error", err)
		httpx.Write(w, http.StatusInternalServerError, "failed to initiate payment", nil)
		return
	}

	httpx.Write(w, http.StatusCreated, "payment initiated successfully", payment)
}

func (h *Handler) ConfirmPayment(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r)
	if err != nil {
		httpx.Write(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	paymentID := r.PathValue("id")
	if paymentID == "" {
		httpx.Write(w, http.StatusBadRequest, "payment id required", nil)
		return
	}

	err = h.service.ConfirmPayment(r.Context(), paymentID, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httpx.Write(w, http.StatusNotFound, "payment not found", nil)
			return
		}
		if errors.Is(err, ErrUnauthorized) {
			httpx.Write(w, http.StatusForbidden, "unauthorized", nil)
			return
		}
		if errors.Is(err, database.ErrPaymentNotPending) {
			httpx.Write(w, http.StatusConflict, "payment is not pending", nil)
			return
		}
		slog.Error("failed to confirm payment", "error", err)
		httpx.Write(w, http.StatusInternalServerError, "failed to confirm payment", nil)
		return
	}

	httpx.Write(w, http.StatusOK, "payment confirmed successfully", nil)
}

func (h *Handler) CancelPayment(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r)
	if err != nil {
		httpx.Write(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	paymentID := r.PathValue("id")
	if paymentID == "" {
		httpx.Write(w, http.StatusBadRequest, "payment id required", nil)
		return
	}

	err = h.service.CancelPayment(r.Context(), paymentID, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httpx.Write(w, http.StatusNotFound, "payment not found", nil)
			return
		}
		if errors.Is(err, ErrUnauthorized) {
			httpx.Write(w, http.StatusForbidden, "unauthorized", nil)
			return
		}
		if errors.Is(err, database.ErrNoItem) || errors.Is(err, pgx.ErrNoRows) {
			httpx.Write(w, http.StatusBadRequest, "order is not pending", nil)
			return
		}
		if errors.Is(err, database.ErrPaymentNotPending) {
			httpx.Write(w, http.StatusConflict, "payment is not pending", nil)
			return
		}
		slog.Error("failed to cancel payment", "error", err)
		httpx.Write(w, http.StatusInternalServerError, "failed to cancel payment", nil)
		return
	}

	httpx.Write(w, http.StatusOK, "payment cancelled successfully", nil)
}

func (h *Handler) GetPaymentByID(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r)
	if err != nil {
		httpx.Write(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	paymentID := r.PathValue("id")
	if paymentID == "" {
		httpx.Write(w, http.StatusBadRequest, "payment id required", nil)
		return
	}

	payment, err := h.service.GetPaymentByID(r.Context(), paymentID, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httpx.Write(w, http.StatusNotFound, "payment not found", nil)
			return
		}
		if errors.Is(err, ErrUnauthorized) {
			httpx.Write(w, http.StatusForbidden, "unauthorized", nil)
			return
		}
		slog.Error("failed to get payment", "error", err)
		httpx.Write(w, http.StatusInternalServerError, "failed to fetch payment", nil)
		return
	}

	httpx.Write(w, http.StatusOK, "payment fetched successfully", payment)
}
