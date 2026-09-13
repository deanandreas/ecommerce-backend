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
	GetUserCart(ctx context.Context, userID string) (db.GetUserCartRow, error)
	GetAllUserOrders(ctx context.Context, userID string) ([]db.GetAllUserOrdersRow, error)
	GetUserOrder(ctx context.Context, arg db.GetUserOrderParams) (db.GetUserOrderRow, error)
	GetUnpayedOrder(ctx context.Context, userID string) (db.GetUnpayedOrderRow, error)
	GetPaymentByID(ctx context.Context, paymentID string) (db.Payment, error)
	GetUserReview(ctx context.Context, arg db.GetUserReviewParams) (db.GetUserReviewRow, error)
	GetProductReviews(ctx context.Context, productID string) ([]db.GetProductReviewsRow, error)
	// >> Create Logics
	InsertCartTx(ctx context.Context, arg database.CreateCart) (*db.CartItem, error)
	InserOrderTx(ctx context.Context, arg db.InsertOrderParams) (*db.GetUserOrderRow, error)
	InsertPayment(ctx context.Context, arg db.InsertPaymentParams) (db.Payment, error)
	InsertReviewTx(ctx context.Context, arg db.InsertReviewParams) (*db.GetReviewByIDRow, error)
	// >> Update Logics
	UpdateCartItemTx(ctx context.Context, arg db.UpdateCartItemQuantityParams) (*db.CartItem, error)
	ConfirmPaymentTx(ctx context.Context, paymentID, orderID, userID string) error
	UpdateAddress(ctx context.Context, arg db.UpdateAddressParams) error
	CancelPaymentTx(ctx context.Context, paymentID, orderID, userID string) error
	// >> Delete Logics
	DeleteUser(ctx context.Context, userID string) error
	DeleteAllCartItem(ctx context.Context, arg db.DeleteAllCartItemParams) (int64, error)
	DeleteCartItem(ctx context.Context, arg db.DeleteCartItemParams) (int64, error)
	// >> Other Logics
	Close()
	Health(ctx context.Context) map[string]string
}

type Server struct {
	port int
	http *http.Server
	db   DBService
	*Handlers
}

func NewAPI() (*Server, error) {
	port, dbURL, err := GetENV()
	if err != nil {
		return nil, err
	}
	pool, handlers, err := Handler(dbURL)
	if err != nil {
		return nil, err
	}

	app := &Server{
		port:     port,
		db:       pool,
		Handlers: handlers,
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
