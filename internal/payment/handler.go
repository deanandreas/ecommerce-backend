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
		httpx.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}

	orderID := r.PathValue("id")
	if orderID == "" {
		httpx.Error(w, http.StatusBadRequest, "ORDER_ID_REQUIRED", "order id required")
		return
	}

	payment, err := h.service.InitiatePayment(r.Context(), orderID, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httpx.Error(w, http.StatusBadRequest, "ORDER_NOT_FOUND", "no pending order found")
			return
		}
		if errors.Is(err, ErrOrderNotPending) {
			httpx.Error(w, http.StatusBadRequest, "ORDER_NOT_PENDING", "order is not pending")
			return
		}
		slog.Error("failed to initiate payment", "error", err)
		httpx.Error(w, http.StatusInternalServerError, "INTERNAL", "failed to initiate payment")
		return
	}

	httpx.Send(w, http.StatusCreated, payment)
}

func (h *Handler) ConfirmPayment(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r)
	if err != nil {
		httpx.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}

	paymentID := r.PathValue("id")
	if paymentID == "" {
		httpx.Error(w, http.StatusBadRequest, "PAYMENT_ID_REQUIRED", "payment id required")
		return
	}

	err = h.service.ConfirmPayment(r.Context(), paymentID, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httpx.Error(w, http.StatusNotFound, "PAYMENT_NOT_FOUND", "payment not found")
			return
		}
		if errors.Is(err, ErrUnauthorized) {
			httpx.Error(w, http.StatusForbidden, "FORBIDDEN", "unauthorized")
			return
		}
		if errors.Is(err, database.ErrPaymentNotPending) {
			httpx.Error(w, http.StatusConflict, "PAYMENT_NOT_PENDING", "payment is not pending")
			return
		}
		slog.Error("failed to confirm payment", "error", err)
		httpx.Error(w, http.StatusInternalServerError, "INTERNAL", "failed to confirm payment")
		return
	}

	httpx.Send(w, http.StatusNoContent, nil)
}

func (h *Handler) CancelPayment(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r)
	if err != nil {
		httpx.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}

	paymentID := r.PathValue("id")
	if paymentID == "" {
		httpx.Error(w, http.StatusBadRequest, "PAYMENT_ID_REQUIRED", "payment id required")
		return
	}

	err = h.service.CancelPayment(r.Context(), paymentID, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httpx.Error(w, http.StatusNotFound, "PAYMENT_NOT_FOUND", "payment not found")
			return
		}
		if errors.Is(err, ErrUnauthorized) {
			httpx.Error(w, http.StatusForbidden, "FORBIDDEN", "unauthorized")
			return
		}
		if errors.Is(err, database.ErrNoItem) || errors.Is(err, pgx.ErrNoRows) {
			httpx.Error(w, http.StatusBadRequest, "ORDER_NOT_PENDING", "order is not pending")
			return
		}
		if errors.Is(err, database.ErrPaymentNotPending) {
			httpx.Error(w, http.StatusConflict, "PAYMENT_NOT_PENDING", "payment is not pending")
			return
		}
		slog.Error("failed to cancel payment", "error", err)
		httpx.Error(w, http.StatusInternalServerError, "INTERNAL", "failed to cancel payment")
		return
	}

	httpx.Send(w, http.StatusNoContent, nil)
}

func (h *Handler) GetPaymentByID(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r)
	if err != nil {
		httpx.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}

	paymentID := r.PathValue("id")
	if paymentID == "" {
		httpx.Error(w, http.StatusBadRequest, "PAYMENT_ID_REQUIRED", "payment id required")
		return
	}

	payment, err := h.service.GetPaymentByID(r.Context(), paymentID, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httpx.Error(w, http.StatusNotFound, "PAYMENT_NOT_FOUND", "payment not found")
			return
		}
		if errors.Is(err, ErrUnauthorized) {
			httpx.Error(w, http.StatusForbidden, "FORBIDDEN", "unauthorized")
			return
		}
		slog.Error("failed to get payment", "error", err)
		httpx.Error(w, http.StatusInternalServerError, "INTERNAL", "failed to fetch payment")
		return
	}

	httpx.Send(w, http.StatusOK, payment)
}
