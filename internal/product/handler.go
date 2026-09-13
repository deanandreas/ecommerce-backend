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
		httpx.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized")
		return
	}

	product, err := h.service.GetUserProducts(r.Context(), userID)
	if err != nil {
		slog.Error("failed to fetch user product", "error", err)
		httpx.Error(w, http.StatusInternalServerError, "INTERNAL", "failed to fetch user product")
		return
	}
	if product == nil {
		product = []db.GetProductsByUserIDRow{}
	}

	httpx.Send(w, http.StatusOK, product)
}

func (h *Handler) GetUserProduct(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r)
	if err != nil {
		httpx.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized")
		return
	}
	id := r.PathValue("id")
	if id == "" {
		httpx.Error(w, http.StatusBadRequest, "PRODUCT_ID_REQUIRED", "product id is required")
		return
	}

	arg := db.GetProductByUserIDParams{ID: id, UserID: userID}
	product, err := h.service.GetUserProduct(r.Context(), arg)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			switch pgErr.Code {
			case pgerrcode.InvalidTextRepresentation:
				httpx.Error(w, http.StatusBadRequest, "INVALID_UUID", "invalid UUID")
				return
			case pgerrcode.NoData:
				httpx.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
				return
			}
		}
		if errors.Is(err, pgx.ErrNoRows) {
			httpx.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
			return
		}
		slog.Error("failed to fetch user product", "error", err)
		httpx.Error(w, http.StatusInternalServerError, "INTERNAL", "failed to fetch user product")
		return
	}

	httpx.Send(w, http.StatusOK, product)
}

func (h *Handler) GetProduct(w http.ResponseWriter, r *http.Request) {
	productID := r.PathValue("id")
	if productID == "" {
		httpx.Error(w, http.StatusBadRequest, "PRODUCT_ID_REQUIRED", "missing product id in path")
		return
	}

	product, err := h.service.GetProduct(r.Context(), productID)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			switch pgErr.Code {
			case pgerrcode.NoData:
				httpx.Error(w, http.StatusNotFound, "PRODUCT_NOT_FOUND", "product does not exist")
				return
			case pgerrcode.InvalidTextRepresentation:
				httpx.Error(w, http.StatusBadRequest, "INVALID_PRODUCT_ID", "invalid product id")
				return
			}
		}
		slog.Error("failed to get product details", "error", err)
		httpx.Error(w, http.StatusInternalServerError, "INTERNAL", "failed to fetch a product")
		return
	}

	httpx.Send(w, http.StatusOK, product)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r)
	if err != nil {
		httpx.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized")
		return
	}

	var req database.ProductData
	if err := httpx.FromData(r, "data", &req); err != nil {
		switch {
		case errors.Is(err, httpx.ErrMissingKey):
			httpx.Error(w, http.StatusBadRequest, "DATA_KEY_REQUIRED", "missing json payload vie data key")
		case errors.As(err, &httpx.ErrTypeUnmarshal):
			httpx.Error(w, http.StatusBadRequest, "INVALID_JSON_VALUE", "invalid type of json payload")
		default:
			httpx.Error(w, http.StatusUnprocessableEntity, "INVALID_JSON", "invalid json request")
		}
		return
	}
	req.UserID = userID

	products, err := h.service.CreateProduct(r.Context(), r, req)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrInvalidData):
			httpx.Error(w, http.StatusBadRequest, "INVALID_PRODUCT_DATA", "invalid product data")
		case errors.Is(err, ErrInvalidType):
			httpx.Error(w, http.StatusBadRequest, "UNSUPPORTED_IMAGE_TYPE", fmt.Sprintf("unsupported image type, only %v", strings.Join(Ext, ", ")))
		case errors.Is(err, ErrNoFile):
			httpx.Error(w, http.StatusBadRequest, "IMAGE_REQUIRED", "images are required")
		case errors.Is(err, http.ErrNotMultipart):
			httpx.Error(w, http.StatusBadRequest, "MULTIPART_REQUIRED", "missing image via multipart/form-data")
		default:
			slog.Error("failed to create product", "error", err)
			httpx.Error(w, http.StatusInternalServerError, "INTERNAL", "failed to create product")
		}
		return
	}

	httpx.Send(w, http.StatusCreated, products)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r)
	if err != nil {
		httpx.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}
	productID := r.PathValue("id")
	if productID == "" {
		httpx.Error(w, http.StatusBadRequest, "PRODUCT_ID_REQUIRED", "product id required")
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
			httpx.Error(w, http.StatusBadRequest, "NO_FIELDS_TO_UPDATE", "no field provided to update product")
			return
		}
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			switch pgErr.Code {
			case pgerrcode.InvalidTextRepresentation:
				httpx.Error(w, http.StatusBadRequest, "INVALID_PRODUCT_ID", "invalid product id")
				return
			}
		}
		slog.Error("failed to update product", "error", err)
		httpx.Error(w, http.StatusInternalServerError, "INTERNAL", "failed to update product")
		return
	}

	httpx.Send(w, http.StatusOK, res)
}

func (h *Handler) AddProductImages(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r)
	if err != nil {
		httpx.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}
	productID := r.PathValue("id")
	if productID == "" {
		httpx.Error(w, http.StatusBadRequest, "PRODUCT_ID_REQUIRED", "product id is required")
		return
	}
	arg := db.GetProductByUserIDParams{ID: productID, UserID: userID}

	product, err := h.service.AddProductImages(r.Context(), r, arg)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidType):
			httpx.Error(w, http.StatusBadRequest, "UNSUPPORTED_IMAGE_TYPE", fmt.Sprintf("unsupported image type, only %v", strings.Join(Ext, ", ")))
		case errors.Is(err, ErrNoFile), errors.Is(err, http.ErrNotMultipart):
			httpx.Error(w, http.StatusBadRequest, "IMAGE_REQUIRED", "images are required via multipart/form-data with image key")
		default:
			slog.Error("failed to insert product images", "error", err)
			httpx.Error(w, http.StatusInternalServerError, "INTERNAL", "failed to save images")
		}
		return
	}

	httpx.Send(w, http.StatusOK, product)
}

func (h *Handler) UpdateDefaultImage(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r)
	if err != nil {
		httpx.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized")
		return
	}

	var req db.GetProductImagesParams
	if !httpx.Read(w, r, &req) {
		return
	}

	if req.ID == "" || req.ProductID == "" {
		httpx.Error(w, http.StatusBadRequest, "IMAGE_ID_PRODUCT_ID_REQUIRED", "image id and product id required")
		return
	}
	arg := db.GetProductByUserIDParams{ID: req.ProductID, UserID: userID}

	product, err := h.service.UpdateDefaultImage(r.Context(), arg, req.ID)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			switch pgErr.Code {
			case pgerrcode.InvalidTextRepresentation:
				httpx.Error(w, http.StatusBadRequest, "INVALID_UUID", "invalid UUID")
				return
			}
		}
		if errors.Is(err, database.ErrNotUpdated) {
			httpx.Error(w, http.StatusBadRequest, "IMAGE_NOT_FOUND", "image does not exist")
			return
		}
		slog.Error("failed to update default image", "error", err)
		httpx.Error(w, http.StatusInternalServerError, "INTERNAL", "failed to change default image")
		return
	}

	httpx.Send(w, http.StatusOK, product)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r)
	if err != nil {
		httpx.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized")
		return
	}
	productID := r.PathValue("id")
	if productID == "" {
		httpx.Error(w, http.StatusBadRequest, "PRODUCT_ID_REQUIRED", "missing product id")
		return
	}
	arg := db.SoftDeleteProductParams{ID: productID, UserID: userID}

	result, err := h.service.DeleteProduct(r.Context(), arg)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			switch pgErr.Code {
			case pgerrcode.InvalidTextRepresentation:
				httpx.Error(w, http.StatusBadRequest, "INVALID_PRODUCT_ID", "invalid product id")
				return
			case pgerrcode.NoData:
				httpx.Error(w, http.StatusNotFound, "PRODUCT_NOT_FOUND", "product does not exist")
				return
			}
		}
		slog.Error("failed to delete product", "error", err)
		httpx.Error(w, http.StatusInternalServerError, "INTERNAL", "failed to delete product")
		return
	}
	if result == 0 {
		httpx.Error(w, http.StatusNotFound, "PRODUCT_NOT_FOUND", "product does not exist")
		return
	}

	httpx.Send(w, http.StatusNoContent, nil)
}

func (h *Handler) DeleteProductImage(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r)
	if err != nil {
		httpx.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized")
		return
	}
	var req db.GetProductImagesParams
	if !httpx.Read(w, r, &req) {
		return
	}

	if req.ID == "" || req.ProductID == "" {
		httpx.Error(w, http.StatusBadRequest, "IMAGE_ID_PRODUCT_ID_REQUIRED", "image id and product id required")
		return
	}

	err = h.service.DeleteProductImage(r.Context(), req, userID)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			switch pgErr.Code {
			case pgerrcode.InvalidTextRepresentation:
				httpx.Error(w, http.StatusBadRequest, "INVALID_UUID", "invalid UUID")
				return
			case pgerrcode.NoData:
				httpx.Error(w, http.StatusNotFound, "PRODUCT_NOT_FOUND", "product does not exist")
				return
			}
		}
		if errors.Is(err, pgx.ErrNoRows) {
			httpx.Error(w, http.StatusBadRequest, "IMAGE_NOT_FOUND", "image does not exist")
			return
		}
		if errors.Is(err, database.ErrNotDeleted) {
			httpx.Error(w, http.StatusBadRequest, "DEFAULT_IMAGE_DELETE_FORBIDDEN", "default image can not be deleted")
			return
		}
		slog.Error("failed to delete product image", "error", err)
		httpx.Error(w, http.StatusInternalServerError, "INTERNAL", "failed to delete product image")
		return
	}

	httpx.Send(w, http.StatusNoContent, nil)
}
