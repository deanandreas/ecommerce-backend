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
		httpx.Write(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	var req db.InsertOrderParams
	if !httpx.Read(w, r, &req) {
		return
	}

	if !req.ShipDate.Valid {
		httpx.Write(w, http.StatusBadRequest, "ship date are required", nil)
		return
	}
	req.UserID = userID

	ctx := r.Context()
	cart, err := h.service.CreateOrder(ctx, req)
	if err != nil {
		if errors.Is(err, ErrInvalidShipDate) {
			httpx.Write(w, http.StatusBadRequest, "invalid date", nil)
			return
		}
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			switch pgErr.Code {
			case pgerrcode.NoData:
				httpx.Write(w, http.StatusBadRequest, "user does not have cart", nil)
				return
			}
		}
		if errors.Is(err, pgx.ErrNoRows) {
			httpx.Write(w, http.StatusBadRequest, "user does not have cart", nil)
			return
		}
		slog.Error("failed to create user order", "error", err)
		httpx.Write(w, http.StatusInternalServerError, "failed to create an order", nil)
		return
	}

	httpx.Write(w, http.StatusCreated, "order created successfully", cart)
}

func (h *Handler) GetAllUserOrders(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r)
	if err != nil {
		httpx.Write(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	ctx := r.Context()
	orders, err := h.service.GetAllUserOrders(ctx, userID)
	if err != nil {
		slog.Error("failed to get user orders", "error", err)
		httpx.Write(w, http.StatusInternalServerError, "failed to fetch user orders", nil)
		return
	}
	if orders == nil {
		orders = []db.GetAllUserOrdersRow{}
	}

	httpx.Write(w, http.StatusOK, "user order fetched successfully", orders)
}

func (h *Handler) GetUserOrder(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r)
	if err != nil {
		httpx.Write(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}
	id := r.PathValue("id")
	if id == "" {
		httpx.Write(w, http.StatusBadRequest, "order id required", nil)
		return
	}

	ctx := r.Context()
	order, err := h.service.GetUserOrder(ctx, db.GetUserOrderParams{UserID: userID, ID: id})
	if err != nil {
		slog.Error("failed to get user order", "error", err)
		httpx.Write(w, http.StatusInternalServerError, "failed to fetch order", nil)
		return
	}

	httpx.Write(w, http.StatusOK, "order fetched successfully", order)
}
