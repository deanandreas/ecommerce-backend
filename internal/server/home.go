package server

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	db "github.com/deanandreas/ecommerce-api/internal/database/sqlc"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

var serveImage = http.FileServer(http.Dir(ProductsImageDir))

const ProductsImageDir = "uploads/products/image"

func (s *Server) GetProducts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	queryVals := r.URL.Query()

	var searchArg, slugArg, sortByArg *string
	if search := queryVals.Get("search"); search != "" {
		searchArg = &search
	}
	if slug := queryVals.Get("slug"); slug != "" {
		slugArg = &slug
	}

	var minPriceArg, maxPriceArg pgtype.Numeric
	if minPrice := queryVals.Get("min_price"); minPrice != "" {
		if val, err := strconv.Atoi(minPrice); err == nil {
			minPriceArg.Scan(val)
		}
	}
	if maxPrice := queryVals.Get("max_price"); maxPrice != "" {
		if val, err := strconv.Atoi(maxPrice); err == nil {
			maxPriceArg.Scan(val)
		}
	}

	if sortBy := queryVals.Get("sort"); sortBy != "" {
		sortByArg = &sortBy
	}

	products, err := s.db.GetProducts(ctx, db.GetProductsParams{Search: searchArg, Slug: slugArg, MinPrice: minPriceArg, MaxPrice: maxPriceArg, SortBy: sortByArg})
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			switch pgErr.Code {
			case pgerrcode.NoData:
				WriteJSON(w, http.StatusNotFound, "data not found", nil)
				return
			}
		}
		slog.Error("failed to get products", "error", err)
		WriteJSON(w, http.StatusInternalServerError, "failed to fetch products", nil)
		return
	}

	if products == nil {
		products = []db.GetProductsRow{}
	}

	WriteJSON(w, http.StatusOK, "products fetched successfully", products)
}

func (s *Server) GetProductsCategory(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("l")
	var limit int
	if limitStr != "" {
		var err error
		limit, err = strconv.Atoi(limitStr)
		if err != nil {
			WriteJSON(w, http.StatusBadRequest, "invalid limit value", nil)
			return
		}
	} else {
		limit = 5
	}

	categories, err := s.db.GetProductsCategory(r.Context(), int32(limit))
	if err != nil {
		slog.Error("failed to get products categories", "error", err)
		WriteJSON(w, http.StatusInternalServerError, "failed to fetch the categories", nil)
		return
	}

	WriteJSON(w, http.StatusOK, "products category fetched successfully", categories)
}

func (s *Server) GetPopularProducts(w http.ResponseWriter, r *http.Request) {
	products, err := s.db.GetPopularProducts(r.Context())
	if err != nil {
		slog.Error("failed to get popular products", "error", err)
		WriteJSON(w, http.StatusInternalServerError, "failed to fetch popular products", nil)
		return
	}
	WriteJSON(w, http.StatusOK, "popular products fetched successfully", products)
}
