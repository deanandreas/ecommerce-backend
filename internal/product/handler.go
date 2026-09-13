package product

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/deanandreas/ecommerce-api/internal/auth"
	"github.com/deanandreas/ecommerce-api/internal/database"
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
	return &Handler{service}
}

func (h *Handler) GetUserProducts(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r)
	if err != nil {
		httpx.Write(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	product, err := h.service.GetUserProducts(r.Context(), userID)
	if err != nil {
		slog.Error("failed to fetch user product", "error", err)
		httpx.Write(w, http.StatusInternalServerError, "failed to fetch user product", nil)
		return
	}
	if product == nil {
		product = []db.GetProductsByUserIDRow{}
	}

	httpx.Write(w, http.StatusOK, "user product fetched successfully", product)
}

func (h *Handler) GetUserProduct(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r)
	if err != nil {
		httpx.Write(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}
	id := r.PathValue("id")
	if id == "" {
		httpx.Write(w, http.StatusBadRequest, "product id is required", nil)
		return
	}

	arg := db.GetProductByUserIDParams{ID: id, UserID: userID}
	product, err := h.service.GetUserProduct(r.Context(), arg)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			switch pgErr.Code {
			case pgerrcode.InvalidTextRepresentation:
				httpx.Write(w, http.StatusBadRequest, "invalid UUID", nil)
				return
			case pgerrcode.NoData:
				httpx.Write(w, http.StatusUnauthorized, "unauthorized", nil)
				return
			}
		}
		if errors.Is(err, pgx.ErrNoRows) {
			httpx.Write(w, http.StatusUnauthorized, "unauthorized", nil)
			return
		}
		slog.Error("failed to fetch user product", "error", err)
		httpx.Write(w, http.StatusInternalServerError, "failed to fetch user product", nil)
		return
	}

	httpx.Write(w, http.StatusOK, "user product fetched successfully", product)
}

func (h *Handler) GetProduct(w http.ResponseWriter, r *http.Request) {
	productID := r.PathValue("id")
	if productID == "" {
		httpx.Write(w, http.StatusBadRequest, "missing product id in path", nil)
		return
	}

	product, err := h.service.GetProduct(r.Context(), productID)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			switch pgErr.Code {
			case pgerrcode.NoData:
				httpx.Write(w, http.StatusNotFound, "product does not exist", nil)
				return
			case pgerrcode.InvalidTextRepresentation:
				httpx.Write(w, http.StatusBadRequest, "invalid product id", nil)
				return
			}
		}
		slog.Error("failed to get product details", "error", err)
		httpx.Write(w, http.StatusInternalServerError, "failed to fetch a product", nil)
		return
	}

	httpx.Write(w, http.StatusOK, "product fetched successfully", product)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r)
	if err != nil {
		httpx.Write(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	var req database.ProductData
	if err := httpx.FromData(r, "data", &req); err != nil {
		switch {
		case errors.Is(err, httpx.ErrMissingKey):
			httpx.Write(w, http.StatusBadRequest, "missing json payload vie data key", nil)
		case errors.As(err, &httpx.ErrTypeUnmarshal):
			httpx.Write(w, http.StatusBadRequest, "invalid type of json payload", nil)
		default:
			httpx.Write(w, http.StatusUnprocessableEntity, "invalid json request", nil)
		}
		return
	}
	req.UserID = userID

	products, err := h.service.CreateProduct(r.Context(), r, req)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrInvalidData):
			httpx.Write(w, http.StatusBadRequest, "invalid product data", nil)
		case errors.Is(err, ErrInvalidType):
			httpx.Write(w, http.StatusBadRequest, fmt.Sprintf("unsupported image type, only %v", strings.Join(Ext, ", ")), nil)
		case errors.Is(err, ErrNoFile):
			httpx.Write(w, http.StatusBadRequest, "images are required", nil)
		case errors.Is(err, http.ErrNotMultipart):
			httpx.Write(w, http.StatusBadRequest, "missing image via multipart/form-data", nil)
		default:
			slog.Error("failed to create product", "error", err)
			httpx.Write(w, http.StatusInternalServerError, "failed to create product", nil)
		}
		return
	}

	httpx.Write(w, http.StatusCreated, "product created successfully", products)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r)
	if err != nil {
		httpx.Write(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}
	productID := r.PathValue("id")
	if productID == "" {
		httpx.Write(w, http.StatusBadRequest, "product id required", nil)
		return
	}
	var req database.UpdateProductData
	if !httpx.Read(w, r, &req) {
		return
	}
	req.UserID = userID
	req.ID = productID

	res, err := h.service.UpdateProduct(r.Context(), req)
	if err != nil {
		if errors.Is(err, ErrNoFiled) {
			httpx.Write(w, http.StatusBadRequest, "no field provided to update product", nil)
			return
		}
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			switch pgErr.Code {
			case pgerrcode.InvalidTextRepresentation:
				httpx.Write(w, http.StatusBadRequest, "invalid product id", nil)
				return
			}
		}
		slog.Error("failed to update product", "error", err)
		httpx.Write(w, http.StatusInternalServerError, "failed to update product", nil)
		return
	}

	httpx.Write(w, http.StatusOK, "product updated successfully", res)
}

func (h *Handler) AddProductImages(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r)
	if err != nil {
		httpx.Write(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}
	productID := r.PathValue("id")
	if productID == "" {
		httpx.Write(w, http.StatusBadRequest, "product id is required", nil)
		return
	}
	arg := db.GetProductByUserIDParams{ID: productID, UserID: userID}

	product, err := h.service.AddProductImages(r.Context(), r, arg)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidType):
			httpx.Write(w, http.StatusBadRequest, fmt.Sprintf("unsupported image type, only %v", strings.Join(Ext, ", ")), nil)
		case errors.Is(err, ErrNoFile), errors.Is(err, http.ErrNotMultipart):
			httpx.Write(w, http.StatusBadRequest, "images are required via multipart/form-data with image key", nil)
		default:
			slog.Error("failed to insert product images", "error", err)
			httpx.Write(w, http.StatusInternalServerError, "failed to save images", nil)
		}
		return
	}

	httpx.Write(w, http.StatusOK, "images update successfully", product)
}

func (h *Handler) UpdateDefaultImage(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r)
	if err != nil {
		httpx.Write(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	var req db.GetProductImagesParams
	if !httpx.Read(w, r, &req) {
		return
	}

	if req.ID == "" || req.ProductID == "" {
		httpx.Write(w, http.StatusBadRequest, "image id and product id required", nil)
		return
	}
	arg := db.GetProductByUserIDParams{ID: req.ProductID, UserID: userID}

	product, err := h.service.UpdateDefaultImage(r.Context(), arg, req.ID)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			switch pgErr.Code {
			case pgerrcode.InvalidTextRepresentation:
				httpx.Write(w, http.StatusBadRequest, "invalid UUID", nil)
				return
			}
		}
		if errors.Is(err, database.ErrNotUpdated) {
			httpx.Write(w, http.StatusBadRequest, "image does not exist", nil)
			return
		}
		slog.Error("failed to update default image", "error", err)
		httpx.Write(w, http.StatusInternalServerError, "failed to change default image", nil)
		return
	}

	httpx.Write(w, http.StatusOK, "defual image updated successfully", product)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r)
	if err != nil {
		httpx.Write(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}
	productID := r.PathValue("id")
	if productID == "" {
		httpx.Write(w, http.StatusBadRequest, "missing product id", nil)
		return
	}
	arg := db.SoftDeleteProductParams{ID: productID, UserID: userID}

	result, err := h.service.DeleteProduct(r.Context(), arg)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			switch pgErr.Code {
			case pgerrcode.InvalidTextRepresentation:
				httpx.Write(w, http.StatusBadRequest, "invalid product id", nil)
				return
			case pgerrcode.NoData:
				httpx.Write(w, http.StatusBadRequest, "product does not exist", nil)
				return
			}
		}
		slog.Error("failed to delete product", "error", err)
		httpx.Write(w, http.StatusInternalServerError, "failed to delete product", nil)
		return
	}
	if result == 0 {
		httpx.Write(w, http.StatusBadRequest, "product does not exist", nil)
		return
	}

	httpx.Write(w, http.StatusOK, "product deleted successfully", nil)
}

func (h *Handler) DeleteProductImage(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r)
	if err != nil {
		httpx.Write(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}
	var req db.GetProductImagesParams
	if !httpx.Read(w, r, &req) {
		return
	}

	if req.ID == "" || req.ProductID == "" {
		httpx.Write(w, http.StatusBadRequest, "image id and product id required", nil)
		return
	}

	err = h.service.DeleteProductImage(r.Context(), req, userID)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			switch pgErr.Code {
			case pgerrcode.InvalidTextRepresentation:
				httpx.Write(w, http.StatusBadRequest, "invalid UUID", nil)
				return
			case pgerrcode.NoData:
				httpx.Write(w, http.StatusBadRequest, "product does not exist", nil)
				return
			}
		}
		if errors.Is(err, pgx.ErrNoRows) {
			httpx.Write(w, http.StatusBadRequest, "image does not exist", nil)
			return
		}
		if errors.Is(err, database.ErrNotDeleted) {
			httpx.Write(w, http.StatusBadRequest, "default image can not be deleted", nil)
			return
		}
		slog.Error("failed to delete product image", "error", err)
		httpx.Write(w, http.StatusInternalServerError, "failed to delete product image", nil)
		return
	}

	httpx.Write(w, http.StatusNoContent, "", nil)
}
