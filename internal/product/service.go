package product

import (
	"context"
	"errors"
	"net/http"

	"github.com/deanandreas/ecommerce-api/internal/auth"
	"github.com/deanandreas/ecommerce-api/internal/database"
	db "github.com/deanandreas/ecommerce-api/internal/database/sqlc"
	"github.com/deanandreas/ecommerce-api/internal/storage"
)

var ErrNoFiled = errors.New("no field provided")

type Repository interface {
	GetProductsByUserID(ctx context.Context, userID string) ([]db.GetProductsByUserIDRow, error)
	GetProductByUserID(ctx context.Context, arg db.GetProductByUserIDParams) (db.GetProductByUserIDRow, error)
	GetProductDetails(ctx context.Context, id string) (db.GetProductDetailsRow, error)
	InsertProductTx(ctx context.Context, arg database.ProductData, imageURLs []string) (*db.GetProductByUserIDRow, error)
	InsertProductImagesTx(ctx context.Context, arg db.GetProductByUserIDParams, imageURLs []string) (*db.GetProductByUserIDRow, error)
	UpdateProductTx(ctx context.Context, arg database.UpdateProductData) (*db.GetProductByUserIDRow, error)
	UpdateDefaultImageTx(ctx context.Context, arg db.GetProductByUserIDParams, id string) (*db.GetProductByUserIDRow, error)
	SoftDeleteProduct(ctx context.Context, arg db.SoftDeleteProductParams) (int64, error)
	DeleteProductImageTx(ctx context.Context, arg db.GetProductImagesParams, userID string) (string, error)
}

type Service struct {
	repo  Repository
	store storage.ImageStore
}

func NewService(repo Repository, store storage.ImageStore) *Service {
	return &Service{repo: repo, store: store}
}

func (s *Service) GetUserProducts(ctx context.Context, userID string) ([]db.GetProductsByUserIDRow, error) {
	return s.repo.GetProductsByUserID(ctx, userID)
}

func (s *Service) GetUserProduct(ctx context.Context, arg db.GetProductByUserIDParams) (db.GetProductByUserIDRow, error) {
	return s.repo.GetProductByUserID(ctx, arg)
}

func (s *Service) GetProduct(ctx context.Context, id string) (db.GetProductDetailsRow, error) {
	return s.repo.GetProductDetails(ctx, id)
}

func (s *Service) CreateProduct(ctx context.Context, r *http.Request, arg database.ProductData) (*db.GetProductByUserIDRow, error) {
	if arg.Category.Name == "" || arg.Title == "" || arg.PriceInCent <= 0 || arg.Stock <= 0 || arg.UserID == "" {
		return nil, auth.ErrInvalidData
	}

	imageURLs, err := saveImages(ctx, s.store, r)
	if err != nil {
		return nil, err
	}

	product, err := s.repo.InsertProductTx(ctx, arg, imageURLs)
	if err != nil {
		deleteFiles(ctx, s.store, imageURLs)
		return nil, err
	}

	return product, nil
}

func (s *Service) UpdateProduct(ctx context.Context, arg database.UpdateProductData) (*db.GetProductByUserIDRow, error) {
	if arg.Title == nil && arg.Description == nil && arg.Stock == nil && arg.PriceInCent == nil && arg.Category.Name == "" {
		return nil, ErrNoFiled
	}

	return s.repo.UpdateProductTx(ctx, arg)
}

func (s *Service) AddProductImages(ctx context.Context, r *http.Request, arg db.GetProductByUserIDParams) (*db.GetProductByUserIDRow, error) {
	imageURLs, err := saveImages(ctx, s.store, r)
	if err != nil {
		return nil, err
	}

	product, err := s.repo.InsertProductImagesTx(ctx, arg, imageURLs)
	if err != nil {
		deleteFiles(ctx, s.store, imageURLs)
		return nil, err
	}

	return product, nil
}

func (s *Service) UpdateDefaultImage(ctx context.Context, arg db.GetProductByUserIDParams, id string) (*db.GetProductByUserIDRow, error) {
	return s.repo.UpdateDefaultImageTx(ctx, arg, id)
}

func (s *Service) DeleteProduct(ctx context.Context, arg db.SoftDeleteProductParams) (int64, error) {
	return s.repo.SoftDeleteProduct(ctx, arg)
}

func (s *Service) DeleteProductImage(ctx context.Context, arg db.GetProductImagesParams, userID string) error {
	imageURL, err := s.repo.DeleteProductImageTx(ctx, arg, userID)
	if err != nil {
		return err
	}

	deleteFiles(ctx, s.store, []string{imageURL})
	return nil
}
