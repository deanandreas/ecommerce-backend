package server

import (
	"net/http"

	"github.com/deanandreas/ecommerce-api/internal/storage"
)

func (s *Server) RegisterRoutes() http.Handler {
	mux := http.NewServeMux()

	// -->> System <<-
	mux.HandleFunc("/api/v1/", s.System.Greating)
	mux.HandleFunc("GET /api/v1/health", s.System.GetHealth)

	// -->> Auth <<--
	mux.HandleFunc("POST /api/v1/register", s.Auth.Register)
	mux.HandleFunc("GET /api/v1/token/refresh", s.Auth.Refresh)
	mux.HandleFunc("POST /api/v1/login", s.Auth.Login)

	// -->> Home <<--
	mux.HandleFunc("GET /api/v1/home", s.Home.Home)
	mux.HandleFunc("GET /api/v1/product/star", s.Home.Popular)
	mux.HandleFunc("GET /api/v1/product/details/{id}", s.Product.GetProduct)
	mux.Handle("GET /api/v1/products/image/", http.StripPrefix("/api/v1/products/image/", storage.ImageHandler(s.Store)))

	// -->>  User <<--
	mux.Handle("GET /api/v1/user/profiles",
		s.Middleware.Auth(http.HandlerFunc(s.User.GetUserProfile)))
	mux.Handle("PATCH /api/v1/user/profiles",
		s.Middleware.Auth(http.HandlerFunc(s.User.UpdateUserProfile)))
	mux.Handle("POST /api/v1/user/address",
		s.Middleware.Auth(http.HandlerFunc(s.User.AddUserAddress)))
	mux.Handle("PATCH /api/v1/user/address/{id}",
		s.Middleware.Auth(http.HandlerFunc(s.User.UpdateDefaultAddress)))
	mux.Handle("DELETE /api/v1/user/address/{id}",
		s.Middleware.Auth(http.HandlerFunc(s.User.DeleteAddress)))

	// -->>  Product <<--
	mux.Handle("GET /api/v1/user/products",
		s.Middleware.Auth(http.HandlerFunc(s.Product.GetUserProducts)))
	mux.Handle("GET /api/v1/user/product/{id}",
		s.Middleware.Auth(http.HandlerFunc(s.Product.GetUserProduct)))
	mux.Handle("POST /api/v1/user/product/images/{id}",
		s.Middleware.Auth(http.HandlerFunc(s.Product.AddProductImages)))
	mux.Handle("POST /api/v1/user/products",
		s.Middleware.Auth(http.HandlerFunc(s.Product.Create)))
	mux.Handle("PATCH /api/v1/user/products/{id}",
		s.Middleware.Auth(http.HandlerFunc(s.Product.Update)))
	mux.Handle("PATCH /api/v1/user/product/images",
		s.Middleware.Auth(http.HandlerFunc(s.Product.UpdateDefaultImage)))
	mux.Handle("DELETE /api/v1/user/product/images",
		s.Middleware.Auth(http.HandlerFunc(s.Product.DeleteProductImage)))
	mux.Handle("DELETE /api/v1/user/products/{id}",
		s.Middleware.Auth(http.HandlerFunc(s.Product.Delete)))

	// -->> Cart <<--
	mux.Handle("POST /api/v1/user/carts",
		s.Middleware.Auth(http.HandlerFunc(s.Cart.CreateCart)))
	mux.Handle("GET /api/v1/user/cart/items",
		s.Middleware.Auth(http.HandlerFunc(s.Cart.GetUserCarts)))
	mux.Handle("PATCH /api/v1/user/cart/items",
		s.Middleware.Auth(http.HandlerFunc(s.Cart.UpdateCartQuantity)))
	mux.Handle("DELETE /api/v1/user/cart/items/{id}",
		s.Middleware.Auth(http.HandlerFunc(s.Cart.DeleteCart)))
	mux.Handle("DELETE /api/v1/user/carts/{id}",
		s.Middleware.Auth(http.HandlerFunc(s.Cart.DeleteCarts)))

	// -->> Order <<--
	mux.Handle("POST /api/v1/user/orders",
		s.Middleware.Auth(http.HandlerFunc(s.Order.CreateOrder)))
	mux.Handle("GET /api/v1/user/orders",
		s.Middleware.Auth(http.HandlerFunc(s.Order.GetAllUserOrders)))
	mux.Handle("GET /api/v1/user/orders/{id}",
		s.Middleware.Auth(http.HandlerFunc(s.Order.GetUserOrder)))

	// -->> Payment <<--
	mux.Handle("POST /api/v1/user/orders/{id}/payment",
		s.Middleware.Auth(http.HandlerFunc(s.Payment.InitiatePayment)))
	mux.Handle("POST /api/v1/user/payments/{id}/confirm",
		s.Middleware.Auth(http.HandlerFunc(s.Payment.ConfirmPayment)))
	mux.Handle("POST /api/v1/user/payments/{id}/cancel",
		s.Middleware.Auth(http.HandlerFunc(s.Payment.CancelPayment)))
	mux.Handle("GET /api/v1/user/payments/{id}",
		s.Middleware.Auth(http.HandlerFunc(s.Payment.GetPaymentByID)))

	// -->> Reviews <<--
	mux.Handle("POST /api/v1/user/reviews",
		s.Middleware.Auth(http.HandlerFunc(s.CreateReview)))
	mux.Handle("GET /api/v1/user/review/{id}",
		s.Middleware.Auth(http.HandlerFunc(s.GetUserReview)))
	mux.Handle("GET /api/v1/product/review/{id}",
		http.HandlerFunc(s.GetProductReviews))

	return s.Middleware.Logger(s.Middleware.CORS(mux))
}
