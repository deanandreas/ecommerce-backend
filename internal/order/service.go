package order

import (
	"context"
	"errors"
	"math"
	"time"

	db "github.com/deanandreas/ecommerce-api/internal/database/sqlc"
)

var ErrInvalidShipDate = errors.New("invalid ship date")

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateOrder(ctx context.Context, arg db.InsertOrderParams) (*db.GetUserOrderRow, error) {
	if !arg.ShipDate.Valid {
		return nil, ErrInvalidShipDate
	}

	now := time.Now().Truncate(24 * time.Hour)
	targateDate := arg.ShipDate.Time.Truncate(24 * time.Hour)
	daysUntilShip := int(math.Round(targateDate.Sub(now).Hours() / 24))

	switch {
	case daysUntilShip <= 1:
		arg.ShipPriceInCent = 100000
	case daysUntilShip <= 4:
		arg.ShipPriceInCent = 80000
	case daysUntilShip <= 8:
		arg.ShipPriceInCent = 60000
	default:
		return nil, ErrInvalidShipDate
	}

	return s.repo.InsertOrderTx(ctx, arg)
}

func (s *Service) GetAllUserOrders(ctx context.Context, userID string) ([]db.GetAllUserOrdersRow, error) {
	return s.repo.GetAllUserOrders(ctx, userID)
}

func (s *Service) GetUserOrder(ctx context.Context, arg db.GetUserOrderParams) (db.GetUserOrderRow, error) {
	return s.repo.GetUserOrder(ctx, arg)
}
