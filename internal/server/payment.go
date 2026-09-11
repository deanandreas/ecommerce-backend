package server

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/deanandreas/ecommerce-api/internal/database"
	db "github.com/deanandreas/ecommerce-api/internal/database/sqlc"
	"github.com/jackc/pgx/v5"
)

type paymentItem struct {
	PriceInCent int32 `json:"price_in_cent"`
	Quantity    int32 `json:"quantity"`
}

// InitiatePayment creates a pending payment for an order that has not been
// paid yet, and returns the payment details the client needs to pay.
func (s *Server) InitiatePayment(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		WriteJSON(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	orderID := r.PathValue("id")
	if orderID == "" {
		WriteJSON(w, http.StatusBadRequest, "order id required", nil)
		return
	}

	ctx := r.Context()

	order, err := s.db.GetUnpayedOrder(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			WriteJSON(w, http.StatusBadRequest, "no pending order found", nil)
			return
		}
		slog.Error("failed to get pending order", "error", err)
		WriteJSON(w, http.StatusInternalServerError, "failed to initiate payment", nil)
		return
	}

	if order.ID != orderID {
		WriteJSON(w, http.StatusBadRequest, "order is not pending", nil)
		return
	}

	var items []paymentItem
	rowByte, err := json.Marshal(order.OrderItems)
	if err != nil {
		slog.Error("failed to read order items", "error", err)
		WriteJSON(w, http.StatusInternalServerError, "failed to initiate payment", nil)
		return
	}
	if err := json.Unmarshal(rowByte, &items); err != nil {
		slog.Error("failed to read order items", "error", err)
		WriteJSON(w, http.StatusInternalServerError, "failed to initiate payment", nil)
		return
	}

	total := order.ShipPriceInCent
	for _, item := range items {
		total += item.PriceInCent * item.Quantity
	}

	transactionID, err := newTransactionID()
	if err != nil {
		slog.Error("failed to generate transaction id", "error", err)
		WriteJSON(w, http.StatusInternalServerError, "failed to initiate payment", nil)
		return
	}

	payment, err := s.db.InsertPayment(ctx, db.InsertPaymentParams{
		UserID:         userID,
		TransactionID:  transactionID,
		OrderID:        order.ID,
		PaymentGateway: "internal",
		AmountInCent:   total,
		Currency:       "USD",
		PaymentMethod:  "card",
	})
	if err != nil {
		slog.Error("failed to save payment", "error", err)
		WriteJSON(w, http.StatusInternalServerError, "failed to initiate payment", nil)
		return
	}

	WriteJSON(w, http.StatusCreated, "payment initiated successfully", payment)
}

// ConfirmPayment finalizes a previously initiated payment as paid and marks
// the order as confirmed.
func (s *Server) ConfirmPayment(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		WriteJSON(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	paymentID := r.PathValue("id")
	if paymentID == "" {
		WriteJSON(w, http.StatusBadRequest, "payment id required", nil)
		return
	}

	ctx := r.Context()

	payment, err := s.db.GetPaymentByID(ctx, paymentID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			WriteJSON(w, http.StatusNotFound, "payment not found", nil)
			return
		}
		slog.Error("failed to get payment", "error", err)
		WriteJSON(w, http.StatusInternalServerError, "failed to confirm payment", nil)
		return
	}

	if payment.UserID != userID {
		WriteJSON(w, http.StatusForbidden, "unauthorized", nil)
		return
	}

	if err := s.db.ConfirmPaymentTx(ctx, payment.ID, payment.OrderID, userID); err != nil {
		if errors.Is(err, database.ErrPaymentNotPending) {
			WriteJSON(w, http.StatusConflict, "payment is not pending", nil)
			return
		}
		slog.Error("failed to confirm payment", "error", err)
		WriteJSON(w, http.StatusInternalServerError, "failed to confirm payment", nil)
		return
	}

	WriteJSON(w, http.StatusOK, "payment confirmed successfully", nil)
}

// CancelPayment cancels an in-progress payment, cancels the order and returns
// the reserved stock.
func (s *Server) CancelPayment(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		WriteJSON(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	paymentID := r.PathValue("id")
	if paymentID == "" {
		WriteJSON(w, http.StatusBadRequest, "payment id required", nil)
		return
	}

	ctx := r.Context()

	payment, err := s.db.GetPaymentByID(ctx, paymentID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			WriteJSON(w, http.StatusNotFound, "payment not found", nil)
			return
		}
		slog.Error("failed to get payment", "error", err)
		WriteJSON(w, http.StatusInternalServerError, "failed to cancel payment", nil)
		return
	}

	if payment.UserID != userID {
		WriteJSON(w, http.StatusForbidden, "unauthorized", nil)
		return
	}

	if err := s.db.CancelPaymentTx(ctx, payment.ID, payment.OrderID, userID); err != nil {
		if errors.Is(err, database.ErrNoItem) || errors.Is(err, pgx.ErrNoRows) {
			WriteJSON(w, http.StatusBadRequest, "order is not pending", nil)
			return
		}
		if errors.Is(err, database.ErrPaymentNotPending) {
			WriteJSON(w, http.StatusConflict, "payment is not pending", nil)
			return
		}
		slog.Error("failed to cancel payment", "error", err)
		WriteJSON(w, http.StatusInternalServerError, "failed to cancel payment", nil)
		return
	}

	WriteJSON(w, http.StatusOK, "payment cancelled successfully", nil)
}

// GetPaymentByID returns a single payment owned by the authenticated user.
func (s *Server) GetPaymentByID(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		WriteJSON(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	paymentID := r.PathValue("id")
	if paymentID == "" {
		WriteJSON(w, http.StatusBadRequest, "payment id required", nil)
		return
	}

	payment, err := s.db.GetPaymentByID(r.Context(), paymentID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			WriteJSON(w, http.StatusNotFound, "payment not found", nil)
			return
		}
		slog.Error("failed to get payment", "error", err)
		WriteJSON(w, http.StatusInternalServerError, "failed to fetch payment", nil)
		return
	}

	if payment.UserID != userID {
		WriteJSON(w, http.StatusForbidden, "unauthorized", nil)
		return
	}

	WriteJSON(w, http.StatusOK, "payment fetched successfully", payment)
}

func newTransactionID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
