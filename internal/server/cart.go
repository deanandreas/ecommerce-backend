package server

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/deanandreas/ecommerce-api/internal/database"
	db "github.com/deanandreas/ecommerce-api/internal/database/sqlc"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

func (s *Server) CreateCart(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		WriteJSON(w, http.StatusBadRequest, "missing user id", nil)
		return
	}

	var req database.CreateCart
	if ok := ReadJSON(w, r, &req); !ok {
		return
	}

	if req.ProductID == "" || req.Quantity <= 0 {
		WriteJSON(w, http.StatusBadRequest, "product id and quantity are requirerd", nil)
		return
	}
	ctx := r.Context()
	req.UserID = userID

	cart, err := s.db.InsertCartTx(ctx, req)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			switch pgErr.Code {
			case pgerrcode.ForeignKeyViolation:
				WriteJSON(w, http.StatusBadRequest, "product does not exist", nil)
				return
			case pgerrcode.InvalidTextRepresentation:
				WriteJSON(w, http.StatusBadRequest, "invalid product UUID", nil)
				return
			}
		}
		if errors.Is(err, database.ErrInvalidQuantity) {
			WriteJSON(w, http.StatusBadRequest, "to many quantity", nil)
			return
		}
		slog.Error("failed to create user cart", "error", err)
		WriteJSON(w, http.StatusInternalServerError, "failed to save cart in database", nil)
		return
	}

	WriteJSON(w, http.StatusCreated, "cart created successfully", cart)
}

func (s *Server) UpdateCartQuantity(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		WriteJSON(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}
	var req db.UpdateCartItemQuantityParams
	if ok := ReadJSON(w, r, &req); !ok {
		return
	}

	if req.Quantity <= 0 || req.ID == "" {
		WriteJSON(w, http.StatusBadRequest, "cart id, product id, quantity and item id are required", nil)
		return
	}
	req.UserID = userID

	cart, err := s.db.UpdateCartItemTx(r.Context(), req)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			switch pgErr.Code {
			case pgerrcode.NoData:
				WriteJSON(w, http.StatusBadRequest, "product does not exist", nil)
				return
			case pgerrcode.InvalidTextRepresentation:
				WriteJSON(w, http.StatusBadRequest, "invalid UUID value", nil)
				return
			case pgerrcode.ForeignKeyViolation:
				WriteJSON(w, http.StatusForbidden, "unauthorized", nil)
				return
			}
		}
		if errors.Is(err, database.ErrInvalidQuantity) {
			WriteJSON(w, http.StatusBadRequest, "to many quantity", nil)
			return
		}
		if errors.Is(err, database.ErrNotUpdated) {
			WriteJSON(w, http.StatusBadRequest, "item does not exist", nil)
			return
		}
		slog.Error("failed to update user cart quantity", "error", err)
		WriteJSON(w, http.StatusInternalServerError, "failed to update cart item", nil)
		return
	}

	WriteJSON(w, http.StatusOK, "item quantity updated successfully", cart)
}

func (s *Server) GetUserCarts(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		WriteJSON(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}
	ctx := r.Context()

	cart, err := s.db.GetUserCart(ctx, userID)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			switch pgErr.Code {
			case pgerrcode.NoData:
				WriteJSON(w, http.StatusOK, "user data fetched successfully", cart)
				return
			case pgerrcode.ForeignKeyViolation:
			}
		}
		slog.Error("failed to get user cart", "error", err)
		WriteJSON(w, http.StatusInternalServerError, "failed to fech user cart", nil)
		return
	}

	WriteJSON(w, http.StatusOK, "user cart feched successfully", cart)
}

func (s *Server) DeleteCart(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		WriteJSON(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}
	itemID := r.PathValue("id")
	if itemID == "" {
		WriteJSON(w, http.StatusBadRequest, "cart item id is required", nil)
		return
	}

	arg := db.DeleteCartItemParams{ID: itemID, UserID: userID}

	rows, err := s.db.DeleteCartItem(r.Context(), arg)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			switch pgErr.Code {
			case pgerrcode.InvalidTextRepresentation:
				WriteJSON(w, http.StatusBadRequest, "invalid item id", nil)
				return
			}
		}
		slog.Error("failed to delete user cart", "error", err)
		WriteJSON(w, http.StatusInternalServerError, "failed to delete cart", nil)
		return
	}
	if rows == 0 {
		WriteJSON(w, http.StatusBadRequest, "item does not exist", nil)
		return
	}

	WriteJSON(w, http.StatusNoContent, "", nil)
}

func (s *Server) DeleteCarts(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		WriteJSON(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}
	cartID := r.PathValue("id")
	if cartID == "" {
		WriteJSON(w, http.StatusBadRequest, "cart id is required", nil)
		return
	}

	arg := db.DeleteAllCartItemParams{CartID: cartID, UserID: userID}

	rows, err := s.db.DeleteAllCartItem(r.Context(), arg)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			switch pgErr.Code {
			case pgerrcode.InvalidTextRepresentation:
				WriteJSON(w, http.StatusBadRequest, "invalid cart id", nil)
				return
			}
		}
		slog.Error("failed to delete user carts", "error", err)
		WriteJSON(w, http.StatusInternalServerError, "failed to delete all cart iteme", nil)
		return
	}
	if rows == 0 {
		WriteJSON(w, http.StatusBadRequest, "cart does not have any item", nil)
		return
	}

	WriteJSON(w, http.StatusNoContent, "", nil)
}
