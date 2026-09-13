package cart

import (
	"context"
	"errors"

	"github.com/deanandreas/ecommerce-api/internal/database"
	db "github.com/deanandreas/ecommerce-api/internal/database/sqlc"
)

var ErrInvalidInput = errors.New("invalid input")

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateCart(ctx context.Context, arg database.CreateCart) (*db.CartItem, error) {
	if arg.ProductID == "" || arg.Quantity <= 0 {
		return nil, ErrInvalidInput
	}

	return s.repo.InsertCartTx(ctx, arg)
}

func (s *Service) UpdateCartQuantity(ctx context.Context, arg db.UpdateCartItemQuantityParams) (*db.CartItem, error) {
	if arg.Quantity <= 0 || arg.ID == "" {
		return nil, ErrInvalidInput
	}

	return s.repo.UpdateCartItemTx(ctx, arg)
}

func (s *Service) GetUserCart(ctx context.Context, userID string) (db.GetUserCartRow, error) {
	return s.repo.GetUserCart(ctx, userID)
}

func (s *Service) DeleteCartItem(ctx context.Context, arg db.DeleteCartItemParams) (int64, error) {
	return s.repo.DeleteCartItem(ctx, arg)
}

func (s *Service) DeleteAllCartItem(ctx context.Context, arg db.DeleteAllCartItemParams) (int64, error) {
	return s.repo.DeleteAllCartItem(ctx, arg)
}
