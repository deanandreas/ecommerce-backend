package order

import (
	"context"

	db "github.com/deanandreas/ecommerce-api/internal/database/sqlc"
)

type Repository interface {
	InsertOrderTx(ctx context.Context, arg db.InsertOrderParams) (*db.GetUserOrderRow, error)
	GetAllUserOrders(ctx context.Context, userID string) ([]db.GetAllUserOrdersRow, error)
	GetUserOrder(ctx context.Context, arg db.GetUserOrderParams) (db.GetUserOrderRow, error)
	GetUnpayedOrder(ctx context.Context, userID string) (db.GetUnpayedOrderRow, error)
}
