package auth

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"time"

	"github.com/deanandreas/ecommerce-api/internal/database"
	db "github.com/deanandreas/ecommerce-api/internal/database/sqlc"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

var (
	ErrInvalidData          = errors.New("invalid data")
	ErrEmailExists          = errors.New("email already exists")
	ErrInvalidCredentials   = errors.New("invalid credentials")
	ErrRefreshTokenNotFound = errors.New("refresh token does not exist")
	ErrRefreshTokenExpired  = errors.New("refresh token is expired")
)

type Repository interface {
	InsertUserTx(ctx context.Context, arg database.UserData) (*db.GetUserByIDRow, error)
	GetUserByEmail(ctx context.Context, email string) (db.GetUserByEmailRow, error)
	InsertRefreshToken(ctx context.Context, arg db.InsertRefreshTokenParams) error
	GetRefreshToken(ctx context.Context, hashToken string) (db.RefreshToken, error)
	DeleteRefreshToken(ctx context.Context, arg db.DeleteRefreshTokenParams) (int64, error)
}

type AuthResult struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
	User         any    `json:"user"`
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Register(ctx context.Context, user database.UserData, rawPassword string) (*AuthResult, error) {
	if user.FullName == "" || user.Phone == "" ||
		len(rawPassword) < 8 || user.Email == "" ||
		!user.Birth.Valid || len(user.Address) == 0 {
		return nil, ErrInvalidData
	}

	addr := user.Address[0]
	if addr.City == "" || addr.Country == "" ||
		addr.State == "" || addr.PostalCode == "" ||
		addr.StreetLine1 == "" {
		return nil, ErrInvalidData
	}
	user.Address[0].IsDefault = true

	if _, err := mail.ParseAddress(user.Email); err != nil {
		return nil, ErrInvalidData
	}

	hash, err := HashPassword(rawPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}
	user.HashPassword = hash

	created, err := s.repo.InsertUserTx(ctx, user)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			switch pgErr.Code {
			case pgerrcode.UniqueViolation:
				return nil, ErrEmailExists
			case pgerrcode.NotNullViolation:
				return nil, ErrInvalidData
			}
		}
		return nil, fmt.Errorf("failed to insert user: %w", err)
	}

	res, err := s.issueTokens(ctx, created.ID)
	if err != nil {
		return nil, err
	}
	res.User = created

	return &res, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (*AuthResult, error) {
	if email == "" || password == "" {
		return nil, ErrInvalidData
	}

	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	if !ValidatePassword(user.HashPassword, password) {
		return nil, ErrInvalidCredentials
	}

	res, err := s.issueTokens(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	res.User = user

	return &res, nil
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (*AuthResult, error) {
	if refreshToken == "" {
		return nil, ErrInvalidData
	}

	hash := HashRefreshToken(refreshToken)

	data, err := s.repo.GetRefreshToken(ctx, hash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRefreshTokenNotFound
		}
		return nil, fmt.Errorf("failed to get refresh token: %w", err)
	}

	if time.Now().After(data.ExpiaredAt.Time) {
		return nil, ErrRefreshTokenExpired
	}

	res, err := s.issueTokens(ctx, data.UserID)
	if err != nil {
		return nil, err
	}

	row, err := s.repo.DeleteRefreshToken(ctx, db.DeleteRefreshTokenParams{ID: data.ID, UserID: data.UserID})
	if err != nil {
		return nil, fmt.Errorf("failed to delete refresh token: %w", err)
	}
	if row == 0 {
		return nil, fmt.Errorf("refresh token was already consumed")
	}

	return &res, nil
}

func (s *Service) issueTokens(ctx context.Context, userID string) (AuthResult, error) {
	token, err := GenerateJWT(userID)
	if err != nil {
		return AuthResult{}, fmt.Errorf("failed to generate access token: %w", err)
	}

	refresh, err := GenerateJWT(userID)
	if err != nil {
		return AuthResult{}, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	exp := time.Now().Add(7 * 24 * time.Hour)
	var expTime pgtype.Date
	if err := expTime.Scan(exp); err != nil {
		return AuthResult{}, fmt.Errorf("failed to scan refresh expiry: %w", err)
	}

	arg := db.InsertRefreshTokenParams{UserID: userID, HashToken: HashRefreshToken(refresh), ExpiaredAt: expTime}
	if err := s.repo.InsertRefreshToken(ctx, arg); err != nil {
		return AuthResult{}, fmt.Errorf("failed to insert refresh token: %w", err)
	}

	return AuthResult{Token: token, RefreshToken: refresh}, nil
}
