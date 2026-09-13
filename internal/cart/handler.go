package cart

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/deanandreas/ecommerce-api/internal/database"
	db "github.com/deanandreas/ecommerce-api/internal/database/sqlc"
	"github.com/deanandreas/ecommerce-api/internal/httpx"
	"github.com/deanandreas/ecommerce-api/internal/middleware"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) CreateCart(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r)
	if err != nil {
		httpx.Write(w, http.StatusBadRequest, "missing user id", nil)
		return
	}

	var req database.CreateCart
	if !httpx.Read(w, r, &req) {
		return
	}
	req.UserID = userID
	ctx := r.Context()

	cart, err := h.service.CreateCart(ctx, req)
	if err != nil {
		if errors.Is(err, ErrInvalidInput) {
			httpx.Write(w, http.StatusBadRequest, "product id and quantity are required", nil)
			return
		}
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			switch pgErr.Code {
			case pgerrcode.ForeignKeyViolation:
				httpx.Write(w, http.StatusBadRequest, "product does not exist", nil)
				return
			case pgerrcode.InvalidTextRepresentation:
				httpx.Write(w, http.StatusBadRequest, "invalid product UUID", nil)
				return
			}
		}
		if errors.Is(err, database.ErrInvalidQuantity) {
			httpx.Write(w, http.StatusBadRequest, "to many quantity", nil)
			return
		}
		slog.Error("failed to create user cart", "error", err)
		httpx.Write(w, http.StatusInternalServerError, "failed to save cart in database", nil)
		return
	}

	httpx.Write(w, http.StatusCreated, "cart created successfully", cart)
}

func (h *Handler) UpdateCartQuantity(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r)
	if err != nil {
		httpx.Write(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}
	var req db.UpdateCartItemQuantityParams
	if !httpx.Read(w, r, &req) {
		return
	}
	req.UserID = userID

	cart, err := h.service.UpdateCartQuantity(r.Context(), req)
	if err != nil {
		if errors.Is(err, ErrInvalidInput) {
			httpx.Write(w, http.StatusBadRequest, "cart id, product id, quantity and item id are required", nil)
			return
		}
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			switch pgErr.Code {
			case pgerrcode.NoData:
				httpx.Write(w, http.StatusBadRequest, "product does not exist", nil)
				return
			case pgerrcode.InvalidTextRepresentation:
				httpx.Write(w, http.StatusBadRequest, "invalid UUID value", nil)
				return
			case pgerrcode.ForeignKeyViolation:
				httpx.Write(w, http.StatusForbidden, "unauthorized", nil)
				return
			}
		}
		if errors.Is(err, database.ErrInvalidQuantity) {
			httpx.Write(w, http.StatusBadRequest, "to many quantity", nil)
			return
		}
		if errors.Is(err, database.ErrNotUpdated) {
			httpx.Write(w, http.StatusBadRequest, "item does not exist", nil)
			return
		}
		slog.Error("failed to update user cart quantity", "error", err)
		httpx.Write(w, http.StatusInternalServerError, "failed to update cart item", nil)
		return
	}

	httpx.Write(w, http.StatusOK, "item quantity updated successfully", cart)
}

func (h *Handler) GetUserCarts(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r)
	if err != nil {
		httpx.Write(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}
	ctx := r.Context()

	cart, err := h.service.GetUserCart(ctx, userID)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			switch pgErr.Code {
			case pgerrcode.NoData:
				httpx.Write(w, http.StatusOK, "user data fetched successfully", cart)
				return
			case pgerrcode.ForeignKeyViolation:
			}
		}
		slog.Error("failed to get user cart", "error", err)
		httpx.Write(w, http.StatusInternalServerError, "failed to fetch user cart", nil)
		return
	}

	httpx.Write(w, http.StatusOK, "user cart fetched successfully", cart)
}

func (h *Handler) DeleteCart(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r)
	if err != nil {
		httpx.Write(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}
	itemID := r.PathValue("id")
	if itemID == "" {
		httpx.Write(w, http.StatusBadRequest, "cart item id is required", nil)
		return
	}

	arg := db.DeleteCartItemParams{ID: itemID, UserID: userID}

	rows, err := h.service.DeleteCartItem(r.Context(), arg)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			switch pgErr.Code {
			case pgerrcode.InvalidTextRepresentation:
				httpx.Write(w, http.StatusBadRequest, "invalid item id", nil)
				return
			}
		}
		slog.Error("failed to delete user cart", "error", err)
		httpx.Write(w, http.StatusInternalServerError, "failed to delete cart", nil)
		return
	}
	if rows == 0 {
		httpx.Write(w, http.StatusBadRequest, "item does not exist", nil)
		return
	}

	httpx.Write(w, http.StatusNoContent, "", nil)
}

func (h *Handler) DeleteCarts(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r)
	if err != nil {
		httpx.Write(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}
	cartID := r.PathValue("id")
	if cartID == "" {
		httpx.Write(w, http.StatusBadRequest, "cart id is required", nil)
		return
	}

	arg := db.DeleteAllCartItemParams{CartID: cartID, UserID: userID}

	rows, err := h.service.DeleteAllCartItem(r.Context(), arg)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			switch pgErr.Code {
			case pgerrcode.InvalidTextRepresentation:
				httpx.Write(w, http.StatusBadRequest, "invalid cart id", nil)
				return
			}
		}
		slog.Error("failed to delete user carts", "error", err)
		httpx.Write(w, http.StatusInternalServerError, "failed to delete all cart iteme", nil)
		return
	}
	if rows == 0 {
		httpx.Write(w, http.StatusBadRequest, "cart does not have any item", nil)
		return
	}

	httpx.Write(w, http.StatusNoContent, "", nil)
}
