package payment

import (
	"context"

	db "github.com/deanandreas/ecommerce-api/internal/database/sqlc"
)

type Repository interface {
	GetUnpayedOrder(ctx context.Context, userID string) (db.GetUnpayedOrderRow, error)
	InsertPayment(ctx context.Context, arg db.InsertPaymentParams) (db.Payment, error)
	GetPaymentByID(ctx context.Context, paymentID string) (db.Payment, error)
	ConfirmPaymentTx(ctx context.Context, paymentID, orderID, userID string) error
	CancelPaymentTx(ctx context.Context, paymentID, orderID, userID string) error
}
