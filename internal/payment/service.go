package payment

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"

	db "github.com/deanandreas/ecommerce-api/internal/database/sqlc"
)

var (
	ErrOrderNotPending = errors.New("order is not pending")
	ErrUnauthorized    = errors.New("unauthorized")
)

type paymentItem struct {
	PriceInCent int32 `json:"price_in_cent"`
	Quantity    int32 `json:"quantity"`
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// InitiatePayment creates a pending payment for an order that has not been
// paid yet, and returns the payment details the client needs to pay.
func (s *Service) InitiatePayment(ctx context.Context, orderID, userID string) (db.Payment, error) {
	order, err := s.repo.GetUnpayedOrder(ctx, userID)
	if err != nil {
		return db.Payment{}, err
	}

	if order.ID != orderID {
		return db.Payment{}, ErrOrderNotPending
	}

	var items []paymentItem
	rowByte, err := json.Marshal(order.OrderItems)
	if err != nil {
		return db.Payment{}, err
	}
	if err := json.Unmarshal(rowByte, &items); err != nil {
		return db.Payment{}, err
	}

	total := order.ShipPriceInCent
	for _, item := range items {
		total += item.PriceInCent * item.Quantity
	}

	transactionID, err := newTransactionID()
	if err != nil {
		return db.Payment{}, err
	}

	return s.repo.InsertPayment(ctx, db.InsertPaymentParams{
		UserID:         userID,
		TransactionID:  transactionID,
		OrderID:        order.ID,
		PaymentGateway: "internal",
		AmountInCent:   total,
		Currency:       "USD",
		PaymentMethod:  "card",
	})
}

// ConfirmPayment finalizes a previously initiated payment as paid and marks
// the order as confirmed.
func (s *Service) ConfirmPayment(ctx context.Context, paymentID, userID string) error {
	payment, err := s.repo.GetPaymentByID(ctx, paymentID)
	if err != nil {
		return err
	}

	if payment.UserID != userID {
		return ErrUnauthorized
	}

	return s.repo.ConfirmPaymentTx(ctx, payment.ID, payment.OrderID, userID)
}

// CancelPayment cancels an in-progress payment, cancels the order and returns
// the reserved stock.
func (s *Service) CancelPayment(ctx context.Context, paymentID, userID string) error {
	payment, err := s.repo.GetPaymentByID(ctx, paymentID)
	if err != nil {
		return err
	}

	if payment.UserID != userID {
		return ErrUnauthorized
	}

	return s.repo.CancelPaymentTx(ctx, payment.ID, payment.OrderID, userID)
}

func (s *Service) GetPaymentByID(ctx context.Context, paymentID, userID string) (db.Payment, error) {
	payment, err := s.repo.GetPaymentByID(ctx, paymentID)
	if err != nil {
		return db.Payment{}, err
	}

	if payment.UserID != userID {
		return db.Payment{}, ErrUnauthorized
	}

	return payment, nil
}

func newTransactionID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
