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
		httpx.Error(w, http.StatusNotFound, "NOT_FOUND", "not found")
		return
	}

	httpx.Send(w, http.StatusOK, map[string]string{"greerting": "Hello World!"})
}

func (h *Handler) GetHealth(w http.ResponseWriter, r *http.Request) {
	stats := h.service.Health(r.Context())
	httpx.Send(w, http.StatusOK, stats)
}
