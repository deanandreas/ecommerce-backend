package cart

import (
	"context"

	"github.com/deanandreas/ecommerce-api/internal/database"
	db "github.com/deanandreas/ecommerce-api/internal/database/sqlc"
)

type Repository interface {
	InsertCartTx(ctx context.Context, arg database.CreateCart) (*db.CartItem, error)
	GetUserCart(ctx context.Context, userID string) (db.GetUserCartRow, error)
	UpdateCartItemTx(ctx context.Context, arg db.UpdateCartItemQuantityParams) (*db.CartItem, error)
	DeleteCartItem(ctx context.Context, arg db.DeleteCartItemParams) (int64, error)
	DeleteAllCartItem(ctx context.Context, arg db.DeleteAllCartItemParams) (int64, error)
}
