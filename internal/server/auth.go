package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/deanandreas/ecommerce-api/internal/database"
	db "github.com/deanandreas/ecommerce-api/internal/database/sqlc"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type AuthResponse struct {
	Token string `json:"token"`
	User  any    `json:"user"`
}

func (s *Server) Register(w http.ResponseWriter, r *http.Request) {
	var req database.UserData
	ctx := r.Context()

	if ok := ReadJSON(w, r, &req); !ok {
		return
	}

	if req.FullName == "" || req.Phone == "" ||
		len(req.Password) < 8 || req.Email == "" ||
		!req.Birth.Valid || len(req.Address) == 0 {
		slog.Error("checking", "request", req)
		WriteJSON(w, http.StatusBadRequest, "invalid form of data", nil)
		return
	}

	if req.Address[0].City == "" || req.Address[0].Country == "" ||
		req.Address[0].State == "" || req.Address[0].PostalCode == "" ||
		req.Address[0].StreetLine1 == "" {
		WriteJSON(w, http.StatusBadRequest, "address must contain street_line_1, postal_code, state, country, and city", nil)
		return
	}
	req.Address[0].IsDefault = true

	if _, err := mail.ParseAddress(req.Email); err != nil {
		WriteJSON(w, http.StatusBadRequest, fmt.Sprintf("invalid %v", err), nil)
		return
	}

	var err error
	req.HashPassword, err = HashPassword(req.Password)
	if err != nil {
		slog.Error("hath the password inside create user func", "error", err)
		WriteJSON(w, http.StatusInternalServerError, "failed to register the user", nil)
		return
	}

	user, err := s.db.InsertUserTx(ctx, req)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			switch pgErr.Code {
			case pgerrcode.UniqueViolation:
				WriteJSON(w, http.StatusConflict, "email already in use", nil)
				return
			case pgerrcode.NotNullViolation:
				WriteJSON(w, http.StatusBadRequest, "some field is empty", nil)
				return
			}
		}
		slog.Error("failed to save user to the database", "error", err)
		WriteJSON(w, http.StatusInternalServerError, "failed to register the user", nil)
		return
	}

	token, err := GenerateJWT(user.ID)
	if err != nil {
		slog.Error("failed to generate token", "error", err)
	}

	res := AuthResponse{
		Token: token,
		User:  user,
	}

	s.RefreshTokenUtil(w, ctx, user.ID)

	WriteJSON(w, http.StatusCreated, "user created successfully", res)
}

func (s *Server) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if ok := ReadJSON(w, r, &req); !ok {
		return
	}

	if req.Email == "" || req.Password == "" {
		WriteJSON(w, http.StatusBadRequest, "email and password are requirerd", nil)
		return
	}

	user, err := s.db.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			WriteJSON(w, http.StatusUnauthorized, "invalid email or password", nil)
			return
		}
		slog.Error("get user by email inside login func", "error", err)
		WriteJSON(w, http.StatusInternalServerError, "failed to login", nil)
		return
	}

	ok := ValidatePassword(user.HashPassword, req.Password)
	if !ok {
		WriteJSON(w, http.StatusUnauthorized, "invalid emali or passord", nil)
		return
	}

	token, err := GenerateJWT(user.ID)
	if err != nil {
		slog.Error("GenerateJWT inside Login func", "error", err)
	}

	res := AuthResponse{
		Token: token,
		User:  user,
	}
	s.RefreshTokenUtil(w, ctx, user.ID)
	WriteJSON(w, http.StatusOK, "user login successfully", res)
}

func (s *Server) RefreshTokenUtil(w http.ResponseWriter, ctx context.Context, userID string) {
	cookie, err := GenerateRefreshCookie()
	if err != nil {
		slog.Error("failed to generate refresh token", "error", err)
		return
	}

	hashToken := HashRefreshCookie(*cookie)

	exp := cookie.Expires
	var expTime pgtype.Date
	if err := expTime.Scan(exp); err != nil {
		slog.Error("failed to scan the date", "error", err)
	}
	arg := db.InsertRefreshTokenParams{UserID: userID, HashToken: hashToken, ExpiaredAt: expTime}

	err = s.db.InsertRefreshToken(ctx, arg)

	if err != nil {
		slog.Error("failed to insart refresh token in database", "error", err)
	} else {
		http.SetCookie(w, cookie)
	}
}

func (s *Server) RefreshToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cookie, err := r.Cookie("refresh-token")
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			http.Error(w, "Refresh Token not found", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	hashToken := HashRefreshCookie(*cookie)
	data, err := s.db.GetRefreshToken(ctx, hashToken)
	if err != nil {

		if errors.Is(err, pgx.ErrNoRows) {
			WriteJSON(w, http.StatusUnauthorized, "refresh token does not exist", nil)
			return
		}

		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			switch pgErr.Code {
			case pgerrcode.NoData:
				WriteJSON(w, http.StatusBadRequest, "refresh does not exist", nil)
				return
			}
		}
		slog.Error("failed to get refresh token", "error", err)
		WriteJSON(w, http.StatusInternalServerError, "failed to generated token", nil)
		return
	}

	if time.Now().After(data.ExpiaredAt.Time) {
		WriteJSON(w, http.StatusLocked, "expiared cookie", nil)
		return
	}

	token, err := GenerateJWT(data.UserID)
	if err != nil {
		slog.Error("failed to generate token", "error", err)
	}

	s.RefreshTokenUtil(w, ctx, data.UserID)
	arg := db.DeleteRefreshTokenParams{ID: data.ID, UserID: data.UserID}
	row, err := s.db.DeleteRefreshToken(ctx, arg)
	if err != nil {
		slog.Error("failed to delete refresh	toke", "error", err)
	} else if row == 0 {
		slog.Error("failed to get refresh token", "error", err)
		WriteJSON(w, http.StatusInternalServerError, "failed to generated token", nil)
		return
	}

	WriteJSON(w, http.StatusOK, "token generated successfully", token)
}
