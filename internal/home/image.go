package home

import "net/http"

const ProductsImageDir = "uploads/products/image"

func (h *Handler) ServeImage() http.Handler {
	return http.FileServer(http.Dir(ProductsImageDir))
}
