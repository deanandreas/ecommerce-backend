package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/deanandreas/ecommerce-api/internal/auth"
	db "github.com/deanandreas/ecommerce-api/internal/database/sqlc"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrNoRows          = errors.New("no rows")
	ErrNoFiled         = errors.New("no filed")
	ErrInvalidPassword = errors.New("invalid password")
	ErrInvalidID       = errors.New("invalid id")
)

type Repository interface {
	GetUserByID(ctx context.Context, userID string) (db.GetUserByIDRow, error)
	UpdateProfile(ctx context.Context, arg db.UpdateProfileParams) (db.User, error)
	InsertAddress(ctx context.Context, arg db.InsertAddressParams) error
	UpdateDefaultAddressTx(ctx context.Context, arg db.UpdateDefaultAddressParams) (*db.GetUserByIDRow, error)
	DeleteUserAddress(ctx context.Context, arg db.DeleteUserAddressParams) (int64, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

type UpdateProfile struct {
	db.UpdateProfileParams
	Password string `json:"password"`
}

func (s *Service) UserProfile(ctx context.Context, userID string) (*db.GetUserByIDRow, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNoRows
		}
		return nil, fmt.Errorf("failed to get user by id: %v", err)
	}
	return &user, nil
}

func (s *Service) UpdateProfile(ctx context.Context, profile UpdateProfile) (*db.User, error) {
	if profile.FullName == nil && profile.Password == "" && profile.Phone == nil && !profile.Birth.Valid {
		return nil, ErrNoFiled
	}

	if profile.Password != "" {
		if len(profile.Password) < 8 {
			return nil, ErrInvalidPassword
		}
		hashPassword, err := auth.HashPassword(profile.Password)
		if err != nil {
			return nil, fmt.Errorf("failed to hash user password: %w", err)
		}
		profile.HashPassword = &hashPassword
	}

	user, err := s.repo.UpdateProfile(ctx, profile.UpdateProfileParams)
	if err != nil {
		return nil, fmt.Errorf("failed to update user profile: %w", err)
	}

	return &user, nil
}

func (s *Service) AddAddress(ctx context.Context, addr db.InsertAddressParams) (*db.GetUserByIDRow, error) {
	if addr.StreetLine1 == "" || addr.State == "" || addr.City == "" || addr.PostalCode == "" || addr.Country == "" {
		return nil, auth.ErrInvalidData
	}

	err := s.repo.InsertAddress(ctx, addr)
	if err != nil {
		return nil, fmt.Errorf("failed to insert address: %w", err)
	}

	user, err := s.repo.GetUserByID(ctx, addr.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	return &user, nil
}

func (s *Service) UpdateDefaultAddress(ctx context.Context, addr db.UpdateDefaultAddressParams) (*db.GetUserByIDRow, error) {
	if addr.UserID == "" || addr.ID == "" {
		return nil, auth.ErrInvalidData
	}

	user, err := s.repo.UpdateDefaultAddressTx(ctx, addr)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			switch pgErr.Code {
			case pgerrcode.InvalidTextRepresentation:
				return nil, ErrInvalidID
			}
		}
		return nil, err
	}

	return user, nil
}

func (s *Service) DeleteUserAddress(ctx context.Context, arg db.DeleteUserAddressParams) error {
	if arg.ID == "" || arg.UserID == "" {
		return auth.ErrInvalidData
	}

	row, err := s.repo.DeleteUserAddress(ctx, arg)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			switch pgErr.Code {
			case pgerrcode.InvalidTextRepresentation:
				return ErrInvalidID
			}
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNoRows
		}
		return err
	} else if row == 0 {
		return ErrNoRows
	}

	return nil
}
