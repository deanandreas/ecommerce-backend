package server

import (
	"errors"
	"log/slog"
	"math"
	"net/http"
	"time"

	db "github.com/deanandreas/ecommerce-api/internal/database/sqlc"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func (s *Server) CreateOrder(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		WriteJSON(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	var req db.InsertOrderParams
	if ok := ReadJSON(w, r, &req); !ok {
		return
	}

	if !req.ShipDate.Valid {
		WriteJSON(w, http.StatusBadRequest, "ship date are required", nil)
		return
	}

	now := time.Now().Truncate(24 * time.Hour)
	targateDate := req.ShipDate.Time.Truncate(24 * time.Hour)
	daysUntilShip := int(math.Round(targateDate.Sub(now).Hours() / 24))

	switch {
	case daysUntilShip <= 1:
		req.ShipPriceInCent = 100000
	case daysUntilShip <= 4:
		req.ShipPriceInCent = 80000
	case daysUntilShip <= 8:
		req.ShipPriceInCent = 60000
	default:
		WriteJSON(w, http.StatusBadRequest, "invalid date", nil)
		return
	}
	req.UserID = userID

	ctx := r.Context()
	cart, err := s.db.InserOrderTx(ctx, req)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			switch pgErr.Code {
			case pgerrcode.NoData:
				WriteJSON(w, http.StatusBadRequest, "user does not have cart", nil)
				return
			}
		}
		if errors.Is(err, pgx.ErrNoRows) {
			WriteJSON(w, http.StatusBadRequest, "user does not have cart", nil)
			return
		}
		slog.Error("failed to create user order", "error", err)
		WriteJSON(w, http.StatusInternalServerError, "failed to create an order", nil)
		return
	}

	WriteJSON(w, http.StatusCreated, "order created successfully", cart)
}

func (s *Server) GetAllUserOrders(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		WriteJSON(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	ctx := r.Context()
	orders, err := s.db.GetAllUserOrders(ctx, userID)
	if err != nil {
		slog.Error("failed to get user orders", "error", err)
		WriteJSON(w, http.StatusInternalServerError, "failed to fetch user orders", nil)
		return
	} else if orders == nil {
		orders = []db.GetAllUserOrdersRow{}
	}

	WriteJSON(w, http.StatusOK, "user order fetched successfully", orders)
}

func (s *Server) GetUserOrder(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		WriteJSON(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}
	id := r.PathValue("id")
	if id == "" {
		WriteJSON(w, http.StatusBadRequest, "order id required", nil)
		return
	}

	ctx := r.Context()
	order, err := s.db.GetUserOrder(ctx, db.GetUserOrderParams{UserID: userID, ID: id})
	if err != nil {
		slog.Error("failed to get user order", "error", err)
		WriteJSON(w, http.StatusInternalServerError, "failed to fetch order", nil)
		return
	}

	WriteJSON(w, http.StatusOK, "order fetched successfully", order)
}
