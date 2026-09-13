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
		httpx.Error(w, http.StatusBadRequest, "INVALID_USER_ID", "missing user id")
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
			httpx.Error(w, http.StatusBadRequest, "INVALID_CART_INPUT", "product id and quantity are required")
			return
		}
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			switch pgErr.Code {
			case pgerrcode.ForeignKeyViolation:
				httpx.Error(w, http.StatusBadRequest, "PRODUCT_NOT_FOUND", "product does not exist")
				return
			case pgerrcode.InvalidTextRepresentation:
				httpx.Error(w, http.StatusBadRequest, "INVALID_PRODUCT_ID", "invalid product UUID")
				return
			}
		}
		if errors.Is(err, database.ErrInvalidQuantity) {
			httpx.Error(w, http.StatusBadRequest, "INVALID_QUANTITY", "to many quantity")
			return
		}
		slog.Error("failed to create user cart", "error", err)
		httpx.Error(w, http.StatusInternalServerError, "INTERNAL", "failed to save cart in database")
		return
	}

	httpx.Send(w, http.StatusCreated, cart)
}

func (h *Handler) UpdateCartQuantity(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r)
	if err != nil {
		httpx.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
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
			httpx.Error(w, http.StatusBadRequest, "INVALID_CART_INPUT", "cart id, product id, quantity and item id are required")
			return
		}
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			switch pgErr.Code {
			case pgerrcode.NoData:
				httpx.Error(w, http.StatusBadRequest, "PRODUCT_NOT_FOUND", "product does not exist")
				return
			case pgerrcode.InvalidTextRepresentation:
				httpx.Error(w, http.StatusBadRequest, "INVALID_UUID", "invalid UUID value")
				return
			case pgerrcode.ForeignKeyViolation:
				httpx.Error(w, http.StatusForbidden, "FORBIDDEN", "unauthorized")
				return
			}
		}
		if errors.Is(err, database.ErrInvalidQuantity) {
			httpx.Error(w, http.StatusBadRequest, "INVALID_QUANTITY", "to many quantity")
			return
		}
		if errors.Is(err, database.ErrNotUpdated) {
			httpx.Error(w, http.StatusBadRequest, "CART_ITEM_NOT_FOUND", "item does not exist")
			return
		}
		slog.Error("failed to update user cart quantity", "error", err)
		httpx.Error(w, http.StatusInternalServerError, "INTERNAL", "failed to update cart item")
		return
	}

	httpx.Send(w, http.StatusOK, cart)
}

func (h *Handler) GetUserCarts(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r)
	if err != nil {
		httpx.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}
	ctx := r.Context()

	cart, err := h.service.GetUserCart(ctx, userID)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			switch pgErr.Code {
			case pgerrcode.NoData:
				httpx.Send(w, http.StatusOK, cart)
				return
			case pgerrcode.ForeignKeyViolation:
			}
		}
		slog.Error("failed to get user cart", "error", err)
		httpx.Error(w, http.StatusInternalServerError, "INTERNAL", "failed to fetch user cart")
		return
	}

	httpx.Send(w, http.StatusOK, cart)
}

func (h *Handler) DeleteCart(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r)
	if err != nil {
		httpx.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized")
		return
	}
	itemID := r.PathValue("id")
	if itemID == "" {
		httpx.Error(w, http.StatusBadRequest, "CART_ITEM_ID_REQUIRED", "cart item id is required")
		return
	}

	arg := db.DeleteCartItemParams{ID: itemID, UserID: userID}

	rows, err := h.service.DeleteCartItem(r.Context(), arg)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			switch pgErr.Code {
			case pgerrcode.InvalidTextRepresentation:
				httpx.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid item id")
				return
			}
		}
		slog.Error("failed to delete user cart", "error", err)
		httpx.Error(w, http.StatusInternalServerError, "INTERNAL", "failed to delete cart")
		return
	}
	if rows == 0 {
		httpx.Error(w, http.StatusBadRequest, "CART_ITEM_NOT_FOUND", "item does not exist")
		return
	}

	httpx.Send(w, http.StatusNoContent, nil)
}

func (h *Handler) DeleteCarts(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r)
	if err != nil {
		httpx.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}
	cartID := r.PathValue("id")
	if cartID == "" {
		httpx.Error(w, http.StatusBadRequest, "CART_ID_REQUIRED", "cart id is required")
		return
	}

	arg := db.DeleteAllCartItemParams{CartID: cartID, UserID: userID}

	rows, err := h.service.DeleteAllCartItem(r.Context(), arg)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			switch pgErr.Code {
			case pgerrcode.InvalidTextRepresentation:
				httpx.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid cart id")
				return
			}
		}
		slog.Error("failed to delete user carts", "error", err)
		httpx.Error(w, http.StatusInternalServerError, "INTERNAL", "failed to delete all cart iteme")
		return
	}
	if rows == 0 {
		httpx.Error(w, http.StatusBadRequest, "CART_EMPTY", "cart does not have any item")
		return
	}

	httpx.Send(w, http.StatusNoContent, nil)
}
