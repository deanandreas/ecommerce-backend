package home

import (
	"context"

	db "github.com/deanandreas/ecommerce-api/internal/database/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

type Repository interface {
	GetProducts(ctx context.Context, arg db.GetProductsParams) ([]db.GetProductsRow, error)
	GetProductsCategory(ctx context.Context, limit int32) ([]db.Category, error)
	GetPopularProducts(ctx context.Context) ([]db.GetPopularProductsRow, error)
}

type Request struct {
	Search        *string
	Slug          *string
	SortBy        *string
	MinPrice      pgtype.Numeric
	MaxPrice      pgtype.Numeric
	CategoryLimit int32
}

type Response struct {
	Categories []db.Category       `json:"categories"`
	Products   []db.GetProductsRow `json:"products"`
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Home(ctx context.Context, req Request) (Response, error) {
	if req.Slug != nil && *req.Slug == "all" {
		req.Slug = nil
	}

	if req.CategoryLimit <= 0 {
		req.CategoryLimit = 5
	}

	categories, err := s.repo.GetProductsCategory(ctx, req.CategoryLimit)
	if err != nil {
		return Response{}, err
	}
	if categories == nil {
		categories = []db.Category{}
	}

	allCategory := db.Category{
		ID:   "00000000-0000-0000-0000-000000000000",
		Name: "All",
		Slug: "all",
	}
	categories = append([]db.Category{allCategory}, categories...)

	products, err := s.repo.GetProducts(ctx, db.GetProductsParams{
		Search:   req.Search,
		Slug:     req.Slug,
		MinPrice: req.MinPrice,
		MaxPrice: req.MaxPrice,
		SortBy:   req.SortBy,
	})
	if err != nil {
		return Response{}, err
	}
	if products == nil {
		products = []db.GetProductsRow{}
	}

	return Response{
		Categories: categories,
		Products:   products,
	}, nil
}

func (s *Service) Popular(ctx context.Context) ([]db.GetPopularProductsRow, error) {
	products, err := s.repo.GetPopularProducts(ctx)
	if err != nil {
		return nil, err
	}
	if products == nil {
		products = []db.GetPopularProductsRow{}
	}
	return products, nil
}
