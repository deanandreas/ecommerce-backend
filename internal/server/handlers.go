package server

import (
	"context"

	"github.com/deanandreas/ecommerce-api/internal/auth"
	"github.com/deanandreas/ecommerce-api/internal/cart"
	"github.com/deanandreas/ecommerce-api/internal/database"
	"github.com/deanandreas/ecommerce-api/internal/home"
	"github.com/deanandreas/ecommerce-api/internal/middleware"
	"github.com/deanandreas/ecommerce-api/internal/order"
	"github.com/deanandreas/ecommerce-api/internal/product"
	"github.com/deanandreas/ecommerce-api/internal/storage"
	"github.com/deanandreas/ecommerce-api/internal/system"
	"github.com/deanandreas/ecommerce-api/internal/user"
)

type Handlers struct {
	Home       *home.Handler
	System     *system.Handler
	Auth       *auth.Handler
	User       *user.Handler
	Product    *product.Handler
	Cart       *cart.Handler
	Order      *order.Handler
	Middleware *middleware.Handler
	Store      *storage.MinIOStore
}

func Handler(dbURL string) (DBService, *Handlers, error) {
	ctx := context.Background()
	pool, err := database.GetDB(ctx, dbURL)
	if err != nil {
		return nil, nil, err
	}

	store, err := storage.NewMinIOStore(ctx, GetMinIOConfig())
	if err != nil {
		return nil, nil, err
	}

	handlers := Handlers{
		Home:       home.NewHandler(home.NewService(pool)),
		System:     system.NewHandler(system.NewService(pool)),
		Auth:       auth.NewHandler(auth.NewService(pool)),
		User:       user.NewHandler(user.NewService(pool)),
		Product:    product.NewHandler(product.NewService(pool, store)),
		Cart:       cart.NewHandler(cart.NewService(pool)),
		Order:      order.NewHandler(order.NewService(pool)),
		Middleware: middleware.NewHandler(),
		Store:      store,
	}

	return pool, &handlers, nil
}
