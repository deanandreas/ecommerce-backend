package review

import (
	"context"

	db "github.com/deanandreas/ecommerce-api/internal/database/sqlc"
)

type Repository interface {
	InsertReviewTx(ctx context.Context, arg db.InsertReviewParams) (*db.GetReviewByIDRow, error)
	GetUserReview(ctx context.Context, arg db.GetUserReviewParams) (db.GetUserReviewRow, error)
	GetProductReviews(ctx context.Context, productID string) ([]db.GetProductReviewsRow, error)
	GetLatestReviews(ctx context.Context, limit int32) ([]db.GetLatestReviewsRow, error)
	UpdateReview(ctx context.Context, arg db.UpdateReviewParams) (db.UpdateReviewRow, error)
}
