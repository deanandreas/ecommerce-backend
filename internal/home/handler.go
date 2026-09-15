package home

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/deanandreas/ecommerce-api/internal/httpx"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	queryVals := r.URL.Query()

	var search, slug, sortBy *string
	if v := queryVals.Get("search"); v != "" {
		search = &v
	}
	if v := queryVals.Get("slug"); v != "" {
		slug = &v
	}
	if v := queryVals.Get("sort"); v != "" {
		sortBy = &v
	}

	var minPrice, maxPrice pgtype.Numeric
	if v := queryVals.Get("min_price"); v != "" {
		_ = minPrice.Scan(v)
	}
	if v := queryVals.Get("max_price"); v != "" {
		_ = maxPrice.Scan(v)
	}

	var categoryLimit int32
	if v := queryVals.Get("l"); v != "" {
		if val, err := strconv.Atoi(v); err == nil {
			categoryLimit = int32(val)
		}
	}

	res, err := h.service.Home(ctx, Request{
		Search:        search,
		Slug:          slug,
		SortBy:        sortBy,
		MinPrice:      minPrice,
		MaxPrice:      maxPrice,
		CategoryLimit: categoryLimit,
	})
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			switch pgErr.Code {
			case pgerrcode.NoData:
				httpx.Error(w, http.StatusNotFound, "NOT_FOUND", "data not found")
				return
			}
		}
		slog.Error("failed to get home data", "error", err)
		httpx.Error(w, http.StatusInternalServerError, "INTERNAL", "failed to fetch home data")
		return
	}

	httpx.Send(w, http.StatusOK, res)
}

func (h *Handler) Popular(w http.ResponseWriter, r *http.Request) {
	res, err := h.service.Popular(r.Context())
	if err != nil {
		slog.Error("failed to get popular products", "error", err)
		httpx.Error(w, http.StatusInternalServerError, "INTERNAL", "failed to fetch popular products")
		return
	}

	httpx.Send(w, http.StatusOK, res)
}
