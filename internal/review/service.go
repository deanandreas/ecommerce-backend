package review

import (
	"context"
	"errors"

	db "github.com/deanandreas/ecommerce-api/internal/database/sqlc"
	"github.com/jackc/pgx/v5"
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

	_, err := s.repo.GetUserReview(ctx, db.GetUserReviewParams{UserID: arg.UserID, ProductID: arg.ProductID})
	if err == nil {
		updated, err := s.repo.UpdateReview(ctx, db.UpdateReviewParams{
			UserID:    arg.UserID,
			ProductID: arg.ProductID,
			Rating:    arg.Rating,
			Comment:   arg.Comment,
		})
		if err != nil {
			return nil, err
		}
		return &db.GetReviewByIDRow{ID: updated.ID, Rating: updated.Rating, Comment: updated.Comment}, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	return s.repo.InsertReviewTx(ctx, arg)
}

func (s *Service) GetUserReview(ctx context.Context, arg db.GetUserReviewParams) (db.GetUserReviewRow, error) {
	return s.repo.GetUserReview(ctx, arg)
}

func (s *Service) GetProductReviews(ctx context.Context, productID string) ([]db.GetProductReviewsRow, error) {
	return s.repo.GetProductReviews(ctx, productID)
}

func (s *Service) GetLatestReviews(ctx context.Context, limit int32) ([]db.GetLatestReviewsRow, error) {
	reviews, err := s.repo.GetLatestReviews(ctx, limit)
	if err != nil {
		return nil, err
	}
	if reviews == nil {
		reviews = []db.GetLatestReviewsRow{}
	}
	return reviews, nil
}
