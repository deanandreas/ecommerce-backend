package order

import (
	"errors"
	"log/slog"
	"net/http"

	db "github.com/deanandreas/ecommerce-api/internal/database/sqlc"
	"github.com/deanandreas/ecommerce-api/internal/httpx"
	"github.com/deanandreas/ecommerce-api/internal/middleware"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r)
	if err != nil {
		httpx.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}

	var req db.InsertOrderParams
	if !httpx.Read(w, r, &req) {
		return
	}

	if !req.ShipDate.Valid {
		httpx.Error(w, http.StatusBadRequest, "SHIP_DATE_REQUIRED", "ship date are required")
		return
	}
	req.UserID = userID

	ctx := r.Context()
	cart, err := h.service.CreateOrder(ctx, req)
	if err != nil {
		if errors.Is(err, ErrInvalidShipDate) {
			httpx.Error(w, http.StatusBadRequest, "INVALID_SHIP_DATE", "invalid date")
			return
		}
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			switch pgErr.Code {
			case pgerrcode.NoData:
				httpx.Error(w, http.StatusBadRequest, "CART_NOT_FOUND", "user does not have cart")
				return
			}
		}
		if errors.Is(err, pgx.ErrNoRows) {
			httpx.Error(w, http.StatusBadRequest, "CART_NOT_FOUND", "user does not have cart")
			return
		}
		slog.Error("failed to create user order", "error", err)
		httpx.Error(w, http.StatusInternalServerError, "INTERNAL", "failed to create an order")
		return
	}

	httpx.Send(w, http.StatusCreated, cart)
}

func (h *Handler) GetAllUserOrders(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r)
	if err != nil {
		httpx.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}

	ctx := r.Context()
	orders, err := h.service.GetAllUserOrders(ctx, userID)
	if err != nil {
		slog.Error("failed to get user orders", "error", err)
		httpx.Error(w, http.StatusInternalServerError, "INTERNAL", "failed to fetch user orders")
		return
	}
	if orders == nil {
		orders = []db.GetAllUserOrdersRow{}
	}

	httpx.Send(w, http.StatusOK, orders)
}

func (h *Handler) GetUserOrder(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r)
	if err != nil {
		httpx.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}
	id := r.PathValue("id")
	if id == "" {
		httpx.Error(w, http.StatusBadRequest, "ORDER_ID_REQUIRED", "order id required")
		return
	}

	ctx := r.Context()
	order, err := h.service.GetUserOrder(ctx, db.GetUserOrderParams{UserID: userID, ID: id})
	if err != nil {
		slog.Error("failed to get user order", "error", err)
		httpx.Error(w, http.StatusInternalServerError, "INTERNAL", "failed to fetch order")
		return
	}

	httpx.Send(w, http.StatusOK, order)
}
