package server

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/deanandreas/ecommerce-api/internal/database"
	db "github.com/deanandreas/ecommerce-api/internal/database/sqlc"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func (s *Server) CreateProduct(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	ctx := r.Context()
	if err != nil {
		WriteJSON(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	req := database.ProductData{}
	if err := FromDataToJSON(r, "data", &req); err != nil {
		if errors.Is(err, ErrMissingKey) {
			WriteJSON(w, http.StatusBadRequest, "json payload is required via from-data with data key", nil)
			return
		}
		if errors.As(err, &ErrTypeUnmarshal) {
			WriteJSON(w, http.StatusBadRequest, "invalid type of json payload", nil)
			return
		}
		WriteJSON(w, http.StatusUnprocessableEntity, "invalid json request", nil)
		return
	}

	if req.Category.Name == "" || req.Title == "" || req.PriceInCent <= 0 || req.Stock <= 0 {
		WriteJSON(w, http.StatusBadRequest, "invalid product data", nil)
		return
	}

	req.UserID = userID

	product, err := s.db.InsertProductTx(ctx, r, req)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			switch pgErr.Code {
			case pgerrcode.ForeignKeyViolation:
				WriteJSON(w, http.StatusBadRequest, "user does not exist", nil)
				return
			case pgerrcode.NotNullViolation:
				WriteJSON(w, http.StatusBadRequest, "not null violation", nil)
				return
			case pgerrcode.CheckViolation:
				WriteJSON(w, http.StatusBadRequest, "price_in_cent and stock must not be less than zero", nil)
				return
			case pgerrcode.InvalidTextRepresentation:
				WriteJSON(w, http.StatusBadRequest, "invalid user id", nil)
				return
			}
		}
		if errors.Is(err, database.ErrInvalidType) {
			WriteJSON(w, http.StatusBadRequest, "invalid image file type", nil)
			return
		}
		if errors.Is(err, database.ErrNoFile) {
			WriteJSON(w, http.StatusBadRequest, "images are required", nil)
			return
		}
		if errors.Is(err, http.ErrNotMultipart) {
			WriteJSON(w, http.StatusBadRequest, "missing image via multipart/from-data", nil)
			return
		}
		slog.Error("failed to save product in database", "error", err)
		WriteJSON(w, http.StatusInternalServerError, "failed to create product", nil)
		return
	}

	WriteJSON(w, http.StatusCreated, "product created successfully", product)
}

func (s *Server) GetUserProducts(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		WriteJSON(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	ctx := r.Context()
	product, err := s.db.GetProductsByUserID(ctx, userID)
	if err != nil {
		slog.Error("failed to fetch user product", "error", err)
		WriteJSON(w, http.StatusInternalServerError, "failed to fetch user product", nil)
		return
	}
	if product == nil {
		product = []db.GetProductsByUserIDRow{}
	}

	WriteJSON(w, http.StatusOK, "user product fetched successfully", product)
}

func (s *Server) GetUserProduct(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		WriteJSON(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}
	id := r.PathValue("id")
	if id == "" {
		WriteJSON(w, http.StatusBadRequest, "product id is required", nil)
		return
	}

	arg := db.GetProductByUserIDParams{ID: id, UserID: userID}
	ctx := r.Context()
	product, err := s.db.GetProductByUserID(ctx, arg)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			switch pgErr.Code {
			case pgerrcode.InvalidTextRepresentation:
				WriteJSON(w, http.StatusBadRequest, "invalid UUID", nil)
				return
			case pgerrcode.NoData:
				WriteJSON(w, http.StatusUnauthorized, "unauthorized", nil)
				return
			}
		}
		if errors.Is(err, pgx.ErrNoRows) {
			WriteJSON(w, http.StatusUnauthorized, "unauthorized", nil)
			return
		}
		slog.Error("failed to fetch user product", "error", err)
		WriteJSON(w, http.StatusInternalServerError, "failed to fetch user product", nil)
		return
	}

	WriteJSON(w, http.StatusOK, "user product fetched successfully", product)
}

func (s *Server) GetProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	productID := r.PathValue("id")
	if productID == "" {
		WriteJSON(w, http.StatusBadRequest, "missing product id in path", nil)
		return
	}

	product, err := s.db.GetProductDetails(ctx, productID)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			switch pgErr.Code {
			case pgerrcode.NoData:
				WriteJSON(w, http.StatusNotFound, "product does not exist", nil)
				return
			case pgerrcode.InvalidTextRepresentation:
				WriteJSON(w, http.StatusBadRequest, "invalid product id", nil)
				return
			}
		}
		slog.Error("failed to get product details", "error", err)
		WriteJSON(w, http.StatusInternalServerError, "failed to fetch a product", nil)
		return
	}

	WriteJSON(w, http.StatusOK, "product fetched successfully", product)
}

func (s *Server) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		WriteJSON(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}
	productID := r.PathValue("id")
	if productID == "" {
		WriteJSON(w, http.StatusBadRequest, "product id required", nil)
		return
	}
	var req database.UpdateProductData
	if ok := ReadJSON(w, r, &req); !ok {
		return
	}

	if req.Title == nil && req.Description == nil && req.Stock == nil && req.PriceInCent == nil && req.Category.Name == "" {
		WriteJSON(w, http.StatusBadRequest, "no field provided to update product", nil)
		return
	}
	req.UserID = userID
	req.ID = productID

	ctx := r.Context()
	res, err := s.db.UpdateProductTx(ctx, req)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			switch pgErr.Code {
			case pgerrcode.InvalidTextRepresentation:
				WriteJSON(w, http.StatusBadRequest, "invalid product id", nil)
				return
			}
		}
		slog.Error("failed to update product", "error", err)
		WriteJSON(w, http.StatusInternalServerError, "failed to update product", nil)
		return
	}

	WriteJSON(w, http.StatusOK, "product updated successfully", res)
}

func (s *Server) AddProductImages(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		WriteJSON(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}
	productID := r.PathValue("id")
	if productID == "" {
		WriteJSON(w, http.StatusBadRequest, "product id is required", nil)
		return
	}
	arg := db.GetProductByUserIDParams{ID: productID, UserID: userID}
	ctx := r.Context()
	product, err := s.db.InsertProductImagesTx(ctx, r, arg)
	if err != nil {
		if errors.Is(err, database.ErrInvalidType) {
			WriteJSON(w, http.StatusBadRequest, fmt.Sprintf("unsupported image type, only %v", strings.Join(database.Ext, ", ")), nil)
			return
		}
		if errors.Is(err, http.ErrNotMultipart) {
			WriteJSON(w, http.StatusBadRequest, "images are required via multipart/form-data with image key", nil)
			return
		}
		slog.Error("failed to insert product images", "error", err)
		WriteJSON(w, http.StatusInternalServerError, "failed to save images", nil)
		return
	}

	WriteJSON(w, http.StatusOK, "images update successfully", product)
}

func (s *Server) UpdateDefaultImage(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		WriteJSON(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	var req db.GetProductImagesParams
	if ok := ReadJSON(w, r, &req); !ok {
		return
	}

	if req.ID == "" || req.ProductID == "" {
		WriteJSON(w, http.StatusBadRequest, "image id and product id required", nil)
		return
	}
	arg := db.GetProductByUserIDParams{ID: req.ProductID, UserID: userID}

	ctx := r.Context()
	product, err := s.db.UpdateDefaultImageTx(ctx, arg, req.ID)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			switch pgErr.Code {
			case pgerrcode.InvalidTextRepresentation:
				WriteJSON(w, http.StatusBadRequest, "invalid UUID", nil)
				return
			}
		}
		if errors.Is(err, database.ErrNotUpdated) {
			WriteJSON(w, http.StatusBadRequest, "image does not exist", nil)
			return
		}
		slog.Error("failed to update default image", "error", err)
		WriteJSON(w, http.StatusInternalServerError, "failed to change default image", nil)
		return
	}

	WriteJSON(w, http.StatusOK, "defual image updated successfully", product)
}

func (s *Server) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		WriteJSON(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}
	ctx := r.Context()
	productID := r.PathValue("id")
	if productID == "" {
		WriteJSON(w, http.StatusBadRequest, "missing product id", nil)
		return
	}
	arg := db.SoftDeleteProductParams{ID: productID, UserID: userID}

	result, err := s.db.SoftDeleteProduct(ctx, arg)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			switch pgErr.Code {
			case pgerrcode.InvalidTextRepresentation:
				WriteJSON(w, http.StatusBadRequest, "invalid product id", nil)
				return
			case pgerrcode.NoData:
				WriteJSON(w, http.StatusBadRequest, "product does not exist", nil)
				return
			}
		}
		slog.Error("failed to delete product", "error", err)
		WriteJSON(w, http.StatusInternalServerError, "failed to delete product", nil)
		return
	}
	if result == 0 {
		WriteJSON(w, http.StatusBadRequest, "product does not exist", nil)
		return
	}

	WriteJSON(w, http.StatusOK, "product deleted successfully", nil)
}

func (s *Server) DeleteProductImage(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		WriteJSON(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}
	var req db.GetProductImagesParams
	if ok := ReadJSON(w, r, &req); !ok {
		return
	}

	if req.ID == "" || req.ProductID == "" {
		WriteJSON(w, http.StatusBadRequest, "image id and product id required", nil)
		return
	}

	ctx := r.Context()

	err = s.db.DeleteProductImageTx(ctx, req, userID)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			switch pgErr.Code {
			case pgerrcode.InvalidTextRepresentation:
				WriteJSON(w, http.StatusBadRequest, "invalid UUID", nil)
				return
			case pgerrcode.NoData:
				WriteJSON(w, http.StatusBadRequest, "product does not exist", nil)
				return
			}
		}
		if errors.Is(err, pgx.ErrNoRows) {
			WriteJSON(w, http.StatusBadRequest, "image does not exist", nil)
			return
		}
		if errors.Is(err, database.ErrNotDeleted) {
			WriteJSON(w, http.StatusBadRequest, "default image can not be deleted", nil)
			return
		}
		slog.Error("failed to delete product image", "error", err)
		WriteJSON(w, http.StatusInternalServerError, "failed to delete product image", nil)
		return
	}

	WriteJSON(w, http.StatusNoContent, "", nil)
}
