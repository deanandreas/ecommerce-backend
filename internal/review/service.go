package review

import (
	"context"
	"errors"

	db "github.com/deanandreas/ecommerce-api/internal/database/sqlc"
)

var ErrInvalidInput = errors.New("invalid input")

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateReview(ctx context.Context, arg db.InsertReviewParams) (*db.GetReviewByIDRow, error) {
	if arg.ProductID == "" || arg.Rating == 0 {
		return nil, ErrInvalidInput
	}

	return s.repo.InsertReviewTx(ctx, arg)
}

func (s *Service) GetUserReview(ctx context.Context, arg db.GetUserReviewParams) (db.GetUserReviewRow, error) {
	return s.repo.GetUserReview(ctx, arg)
}

func (s *Service) GetProductReviews(ctx context.Context, productID string) ([]db.GetProductReviewsRow, error) {
	return s.repo.GetProductReviews(ctx, productID)
}
