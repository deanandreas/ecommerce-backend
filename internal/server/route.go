package server

import (
	"net/http"
)

func (s *Server) RegisterRoutes() http.Handler {
	mux := http.NewServeMux()

	// -->> OPEN ROUTES <<-
	mux.HandleFunc("/api/v1/", s.homeHandler)
	mux.HandleFunc("GET /api/v1/health", s.healthHandler)
	mux.HandleFunc("POST /api/v1/register", s.Register)
	mux.HandleFunc("GET /api/v1/token/refresh", s.RefreshToken)
	mux.HandleFunc("POST /api/v1/login", s.Login)
	mux.HandleFunc("GET /api/v1/products", s.GetProducts)
	mux.HandleFunc("GET /api/v1/product/details/{id}", s.GetProduct)

	// -->> NOT OPEN ROUTES <<--
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

	// -->> USER AND CART <<--
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

	// -->> USER AND ORDER <<--
	mux.Handle("POST /api/v1/user/orders",
		s.AuthMiddleware(http.HandlerFunc(s.CreateOrder)))
	mux.Handle("GET /api/v1/user/orders",
		s.AuthMiddleware(http.HandlerFunc(s.GetAllUserOrders)))
	mux.Handle("GET /api/v1/user/orders/{id}",
		s.AuthMiddleware(http.HandlerFunc(s.GetUserOrder)))

	serveImage := http.FileServer(http.Dir(ProductsImageDir))
	mux.Handle("GET /products/image/", http.StripPrefix("/products/image/", serveImage))

	return s.LoggerMiddleware(s.CORSMiddleware(mux))
}
