package server

import (
	"net/http"
)

func (s *Server) RegisterRoutes() http.Handler {
	mux := http.NewServeMux()

	// -->> System <<-
	mux.HandleFunc("/api/v1/", s.homeHandler)
	mux.HandleFunc("GET /api/v1/health", s.healthHandler)

	// -->> Auth <<--
	mux.HandleFunc("POST /api/v1/register", s.Register)
	mux.HandleFunc("GET /api/v1/token/refresh", s.RefreshToken)
	mux.HandleFunc("POST /api/v1/login", s.Login)
	mux.Handle("GET /products/image/", http.StripPrefix("/products/image/", serveImage))

	// -->> Home <<--
	mux.HandleFunc("GET /api/v1/products", s.GetProducts)
	mux.HandleFunc("GET /api/v1/product/categories", s.GetProductsCategory)
	mux.HandleFunc("GET /api/v1/product/star", s.GetPopularProducts)
	mux.HandleFunc("GET /api/v1/product/details/{id}", s.GetProduct)

	// -->>  User <<--
	mux.Handle("GET /api/v1/user/profiles",
		s.AuthMiddleware(http.HandlerFunc(s.GetUserProfile)))
	mux.Handle("PATCH /api/v1/user/profiles",
		s.AuthMiddleware(http.HandlerFunc(s.UpdateUserProfile)))
	mux.Handle("POST /api/v1/user/address",
		s.AuthMiddleware(http.HandlerFunc(s.AddUserAdress)))
	mux.Handle("PATCH /api/v1/user/address/{id}",
		s.AuthMiddleware(http.HandlerFunc(s.UpdateDefaultAddress)))
	mux.Handle("DELETE /api/v1/user/address/{id}",
		s.AuthMiddleware(http.HandlerFunc(s.DeleteUserAddres)))

	// -->>  Product <<--
	mux.Handle("GET /api/v1/user/products",
		s.AuthMiddleware(http.HandlerFunc(s.GetUserProducts)))
	mux.Handle("GET /api/v1/user/product/{id}",
		s.AuthMiddleware(http.HandlerFunc(s.GetUserProduct)))
	mux.Handle("POST /api/v1/user/product/images/{id}",
		s.AuthMiddleware(http.HandlerFunc(s.AddProductImages)))
	mux.Handle("POST /api/v1/user/products",
		s.AuthMiddleware(http.HandlerFunc(s.CreateProduct)))
	mux.Handle("PATCH /api/v1/user/products/{id}",
		s.AuthMiddleware(http.HandlerFunc(s.UpdateProduct)))
	mux.Handle("PATCH /api/v1/user/product/images",
		s.AuthMiddleware(http.HandlerFunc(s.UpdateDefaultImage)))
	mux.Handle("DELETE /api/v1/user/product/images",
		s.AuthMiddleware(http.HandlerFunc(s.DeleteProductImage)))
	mux.Handle("DELETE /api/v1/user/products/{id}",
		s.AuthMiddleware(http.HandlerFunc(s.DeleteProduct)))

	// -->> Cart <<--
	mux.Handle("POST /api/v1/user/carts",
		s.AuthMiddleware(http.HandlerFunc(s.CreateCart)))
	mux.Handle("GET /api/v1/user/cart/items",
		s.AuthMiddleware(http.HandlerFunc(s.GetUserCarts)))
	mux.Handle("PATCH /api/v1/user/cart/items",
		s.AuthMiddleware(http.HandlerFunc(s.UpdateCartQuantity)))
	mux.Handle("DELETE /api/v1/user/cart/items/{id}",
		s.AuthMiddleware(http.HandlerFunc(s.DeleteCart)))
	mux.Handle("DELETE /api/v1/user/carts/{id}",
		s.AuthMiddleware(http.HandlerFunc(s.DeleteCarts)))

	// -->> Order <<--
	mux.Handle("POST /api/v1/user/orders",
		s.AuthMiddleware(http.HandlerFunc(s.CreateOrder)))
	mux.Handle("GET /api/v1/user/orders",
		s.AuthMiddleware(http.HandlerFunc(s.GetAllUserOrders)))
	mux.Handle("GET /api/v1/user/orders/{id}",
		s.AuthMiddleware(http.HandlerFunc(s.GetUserOrder)))

	// -->> Payment <<--
	mux.Handle("POST /api/v1/user/orders/{id}/payment",
		s.AuthMiddleware(http.HandlerFunc(s.InitiatePayment)))
	mux.Handle("POST /api/v1/user/payments/{id}/confirm",
		s.AuthMiddleware(http.HandlerFunc(s.ConfirmPayment)))
	mux.Handle("POST /api/v1/user/payments/{id}/cancel",
		s.AuthMiddleware(http.HandlerFunc(s.CancelPayment)))
	mux.Handle("GET /api/v1/user/payments/{id}",
		s.AuthMiddleware(http.HandlerFunc(s.GetPaymentByID)))

	// -->> Reviews <<--
	mux.Handle("POST /api/v1/user/reviews",
		s.AuthMiddleware(http.HandlerFunc(s.CreateReview)))
	mux.Handle("GET /api/v1/user/review/{id}",
		s.AuthMiddleware(http.HandlerFunc(s.GetUserReview)))
	mux.Handle("GET /api/v1/product/review/{id}",
		http.HandlerFunc(s.GetProductReviews))

	return s.LoggerMiddleware(s.CORSMiddleware(mux))
}
