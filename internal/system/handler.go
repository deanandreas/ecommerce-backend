package system

import (
	"net/http"

	"github.com/deanandreas/ecommerce-api/internal/httpx"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Greating(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/v1/" && r.Method != http.MethodGet {
		httpx.Write(w, http.StatusNotFound, "not found", nil)
		return
	}

	httpx.Write(w, http.StatusOK, "data fetched successfully", map[string]string{"greerting": "Hello World!"})
}

func (h *Handler) GetHealth(w http.ResponseWriter, r *http.Request) {
	stats := h.service.Health(r.Context())
	httpx.Write(w, http.StatusOK, "system health fetched successfully", stats)
}
