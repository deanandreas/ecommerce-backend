package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/deanandreas/ecommerce-api/internal/database"
	db "github.com/deanandreas/ecommerce-api/internal/database/sqlc"
)

type DBService interface {
	// >> Get Logics
	GetUserByID(ctx context.Context, userID string) (db.GetUserByIDRow, error)
	GetUserByEmail(ctx context.Context, email string) (db.GetUserByEmailRow, error)
	GetProducts(ctx context.Context, arg db.GetProductsParams) ([]db.GetProductsRow, error)
	GetUserCart(ctx context.Context, userID string) (db.GetUserCartRow, error)
	GetProductDetails(ctx context.Context, id string) (db.GetProductDetailsRow, error)
	GetProductsByUserID(ctx context.Context, userID string) ([]db.GetProductsByUserIDRow, error)
	GetProductByUserID(ctx context.Context, arg db.GetProductByUserIDParams) (db.GetProductByUserIDRow, error)
	GetRefreshToken(ctx context.Context, hashToken string) (db.RefreshToken, error)
	GetAllUserOrders(ctx context.Context, userID string) ([]db.GetAllUserOrdersRow, error)
	GetUserOrder(ctx context.Context, arg db.GetUserOrderParams) (db.GetUserOrderRow, error)
	GetUnpayedOrder(ctx context.Context, userID string) (db.GetUnpayedOrderRow, error)
	GetPaymentByID(ctx context.Context, paymentID string) (db.Payment, error)
	GetUserReview(ctx context.Context, arg db.GetUserReviewParams) (db.GetUserReviewRow, error)
	GetProductReviews(ctx context.Context, productID string) ([]db.GetProductReviewsRow, error)
	GetProductsCategory(ctx context.Context, limit int32) ([]db.Category, error)
	GetPopularProducts(ctx context.Context) ([]db.GetPopularProductsRow, error)
	// >> Create Logics
	InsertUserTx(ctx context.Context, arg database.UserData) (*db.GetUserByIDRow, error)
	InsertCartTx(ctx context.Context, arg database.CreateCart) (*db.CartItem, error)
	InsertProductTx(ctx context.Context, r *http.Request, arg database.ProductData) (*db.GetProductByUserIDRow, error)
	InsertRefreshToken(ctx context.Context, arg db.InsertRefreshTokenParams) error
	InsertProductImagesTx(ctx context.Context, r *http.Request, arg db.GetProductByUserIDParams) (*db.GetProductByUserIDRow, error)
	InserOrderTx(ctx context.Context, arg db.InsertOrderParams) (*db.GetUserOrderRow, error)
	InsertPayment(ctx context.Context, arg db.InsertPaymentParams) (db.Payment, error)
	InsertAddress(ctx context.Context, arg db.InsertAddressParams) error
	InsertReviewTx(ctx context.Context, arg db.InsertReviewParams) (*db.GetReviewByIDRow, error)
	// >> Update Logics
	UpdateCartItemTx(ctx context.Context, arg db.UpdateCartItemQuantityParams) (*db.CartItem, error)
	UpdateProductTx(ctx context.Context, arg database.UpdateProductData) (*db.GetProductByUserIDRow, error)
	UpdateDefaultImageTx(ctx context.Context, arg db.GetProductByUserIDParams, id string) (*db.GetProductByUserIDRow, error)
	ConfirmPaymentTx(ctx context.Context, paymentID, orderID, userID string) error
	UpdateProfile(ctx context.Context, arg db.UpdateProfileParams) (db.User, error)
	UpdateAddress(ctx context.Context, arg db.UpdateAddressParams) error
	UpdateDefaultAddress(ctx context.Context, arg db.UpdateDefaultAddressParams) (*db.GetUserByIDRow, error)
	CancelPaymentTx(ctx context.Context, paymentID, orderID, userID string) error
	// >> Delete Logics
	DeleteUser(ctx context.Context, userID string) error
	SoftDeleteProduct(ctx context.Context, arg db.SoftDeleteProductParams) (int64, error)
	DeleteAllCartItem(ctx context.Context, arg db.DeleteAllCartItemParams) (int64, error)
	DeleteCartItem(ctx context.Context, arg db.DeleteCartItemParams) (int64, error)
	DeleteRefreshToken(ctx context.Context, arg db.DeleteRefreshTokenParams) (int64, error)
	DeleteProductImageTx(ctx context.Context, arg db.GetProductImagesParams, userID string) error
	DeleteUserAddress(ctx context.Context, arg db.DeleteUserAddressParams) (int64, error)
	// >> Other Logics
	Close()
	Health(ctx context.Context) map[string]string
}

type Server struct {
	port int
	http *http.Server
	db   DBService
}

func NewAPI() (*Server, error) {
	ctx := context.Background()
	port, dbURL, err := GetENV()
	if err != nil {
		return nil, err
	}
	pool, err := database.GetDB(ctx, dbURL)
	if err != nil {
		return nil, err
	}

	app := &Server{
		port: port,
		db:   pool,
	}

	app.http = &http.Server{
		Addr:         fmt.Sprintf(":%d", app.port),
		Handler:      http.TimeoutHandler(app.RegisterRoutes(), 60*time.Second, "server time out"),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 40 * time.Second,
		IdleTimeout:  2 * time.Minute,
	}

	return app, nil
}

func (s *Server) Start() error {
	return s.http.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	err := s.http.Shutdown(ctx)
	s.db.Close()
	return err
}
