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

type HomeResponse struct {
	Categories []db.Category       `json:"categories"`
	Products   []db.GetProductsRow `json:"products"`
}

func (s *Server) Home(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	queryVals := r.URL.Query()

	var searchArg, sortByArg *string
	if search := queryVals.Get("search"); search != "" {
		searchArg = &search
	}

	var slugArg *string
	if slug := queryVals.Get("slug"); slug != "" && slug != "all" {
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

	categoryLimit := int32(5)
	if limitStr := queryVals.Get("l"); limitStr != "" {
		if val, err := strconv.Atoi(limitStr); err == nil && val > 0 {
			categoryLimit = int32(val)
		}
	}

	categories, err := s.db.GetProductsCategory(ctx, categoryLimit)
	if err != nil {
		slog.Error("failed to get categories", "error", err)
		WriteJSON(w, http.StatusInternalServerError, "failed to fetch categories", nil)
		return
	}
	if categories == nil {
		categories = []db.Category{}
	}

	allCategory := db.Category{
		ID:   "00000000-0000-0000-0000-000000000000",
		Name: "All",
		Slug: "all",
	}
	categories = append([]db.Category{allCategory}, categories...)

	products, err := s.db.GetProducts(ctx, db.GetProductsParams{
		Search:   searchArg,
		Slug:     slugArg,
		MinPrice: minPriceArg,
		MaxPrice: maxPriceArg,
		SortBy:   sortByArg,
	})
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

	WriteJSON(w, http.StatusOK, "home data fetched successfully", HomeResponse{
		Categories: categories,
		Products:   products,
	})
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
