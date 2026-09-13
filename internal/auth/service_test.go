package auth

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/deanandreas/ecommerce-api/internal/database"
	db "github.com/deanandreas/ecommerce-api/internal/database/sqlc"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type mockRepo struct {
	insertUser         func(ctx context.Context, arg database.UserData) (*db.GetUserByIDRow, error)
	getUserByEmail     func(ctx context.Context, email string) (db.GetUserByEmailRow, error)
	insertRefreshToken func(ctx context.Context, arg db.InsertRefreshTokenParams) error
	getRefreshToken    func(ctx context.Context, hashToken string) (db.RefreshToken, error)
	deleteRefreshToken func(ctx context.Context, arg db.DeleteRefreshTokenParams) (int64, error)
}

func (m *mockRepo) InsertUserTx(ctx context.Context, arg database.UserData) (*db.GetUserByIDRow, error) {
	return m.insertUser(ctx, arg)
}

func (m *mockRepo) GetUserByEmail(ctx context.Context, email string) (db.GetUserByEmailRow, error) {
	return m.getUserByEmail(ctx, email)
}

func (m *mockRepo) InsertRefreshToken(ctx context.Context, arg db.InsertRefreshTokenParams) error {
	return m.insertRefreshToken(ctx, arg)
}

func (m *mockRepo) GetRefreshToken(ctx context.Context, hashToken string) (db.RefreshToken, error) {
	return m.getRefreshToken(ctx, hashToken)
}

func (m *mockRepo) DeleteRefreshToken(ctx context.Context, arg db.DeleteRefreshTokenParams) (int64, error) {
	return m.deleteRefreshToken(ctx, arg)
}

func TestMain(m *testing.M) {
	os.Setenv("JWT_KEY", "test-secret-key")
	os.Exit(m.Run())
}

func validUserData() database.UserData {
	var birth pgtype.Date
	birth.Scan(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC))

	return database.UserData{
		InsertUserParams: db.InsertUserParams{
			FullName: "Dean Andreas",
			Phone:    "123456789",
			Birth:    birth,
			Email:    "dean@example.com",
		},
		Password: "secret123",
		Address: []db.InsertAddressParams{{
			StreetLine1: "123 Main St",
			PostalCode:  "12345",
			State:       "CA",
			City:        "San Francisco",
			Country:     "USA",
		}},
	}
}

func newService(repo Repository) *Service {
	return NewService(repo)
}

func TestRegisterSuccess(t *testing.T) {
	repo := &mockRepo{
		insertUser: func(ctx context.Context, arg database.UserData) (*db.GetUserByIDRow, error) {
			return &db.GetUserByIDRow{ID: "user-1", Email: arg.Email, FullName: arg.FullName}, nil
		},
		insertRefreshToken: func(ctx context.Context, arg db.InsertRefreshTokenParams) error {
			return nil
		},
	}
	svc := newService(repo)

	res, err := svc.Register(context.Background(), validUserData(), "secret123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res == nil {
		t.Fatal("expected result, got nil")
	}
	if res.Token == "" || res.RefreshToken == "" {
		t.Errorf("expected non-empty tokens, got token=%q refresh=%q", res.Token, res.RefreshToken)
	}
	user, ok := res.User.(*db.GetUserByIDRow)
	if !ok || user.ID != "user-1" {
		t.Errorf("expected created user in result, got %v", res.User)
	}
}

func TestRegisterInvalidData(t *testing.T) {
	repo := &mockRepo{}
	svc := newService(repo)

	user := validUserData()
	user.Password = "short"

	res, err := svc.Register(context.Background(), user, "short")
	if !errors.Is(err, ErrInvalidData) {
		t.Fatalf("expected ErrInvalidData, got %v", err)
	}
	if res != nil {
		t.Errorf("expected nil result, got %v", res)
	}
}

func TestRegisterDuplicateEmail(t *testing.T) {
	repo := &mockRepo{
		insertUser: func(ctx context.Context, arg database.UserData) (*db.GetUserByIDRow, error) {
			return nil, &pgconn.PgError{Code: pgerrcode.UniqueViolation}
		},
	}
	svc := newService(repo)

	_, err := svc.Register(context.Background(), validUserData(), "secret123")
	if !errors.Is(err, ErrEmailExists) {
		t.Fatalf("expected ErrEmailExists, got %v", err)
	}
}

func TestLoginSuccess(t *testing.T) {
	hash, err := HashPassword("secret123")
	if err != nil {
		t.Fatal(err)
	}

	repo := &mockRepo{
		getUserByEmail: func(ctx context.Context, email string) (db.GetUserByEmailRow, error) {
			return db.GetUserByEmailRow{ID: "user-1", Email: email, HashPassword: hash}, nil
		},
		insertRefreshToken: func(ctx context.Context, arg db.InsertRefreshTokenParams) error {
			return nil
		},
	}
	svc := newService(repo)

	res, err := svc.Login(context.Background(), "dean@example.com", "secret123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Token == "" || res.RefreshToken == "" {
		t.Errorf("expected non-empty tokens, got token=%q refresh=%q", res.Token, res.RefreshToken)
	}
}

func TestLoginUnknownUser(t *testing.T) {
	repo := &mockRepo{
		getUserByEmail: func(ctx context.Context, email string) (db.GetUserByEmailRow, error) {
			return db.GetUserByEmailRow{}, pgx.ErrNoRows
		},
	}
	svc := newService(repo)

	_, err := svc.Login(context.Background(), "dean@example.com", "secret123")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestLoginWrongPassword(t *testing.T) {
	hash, err := HashPassword("other-password")
	if err != nil {
		t.Fatal(err)
	}

	repo := &mockRepo{
		getUserByEmail: func(ctx context.Context, email string) (db.GetUserByEmailRow, error) {
			return db.GetUserByEmailRow{ID: "user-1", Email: email, HashPassword: hash}, nil
		},
	}
	svc := newService(repo)

	_, err = svc.Login(context.Background(), "dean@example.com", "secret123")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestRefreshSuccess(t *testing.T) {
	repo := &mockRepo{
		getRefreshToken: func(ctx context.Context, hashToken string) (db.RefreshToken, error) {
			return validRefreshToken(), nil
		},
		insertRefreshToken: func(ctx context.Context, arg db.InsertRefreshTokenParams) error {
			return nil
		},
		deleteRefreshToken: func(ctx context.Context, arg db.DeleteRefreshTokenParams) (int64, error) {
			return 1, nil
		},
	}
	svc := newService(repo)

	res, err := svc.Refresh(context.Background(), "some-refresh-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Token == "" || res.RefreshToken == "" {
		t.Errorf("expected non-empty tokens, got token=%q refresh=%q", res.Token, res.RefreshToken)
	}
}

func TestRefreshTokenNotFound(t *testing.T) {
	repo := &mockRepo{
		getRefreshToken: func(ctx context.Context, hashToken string) (db.RefreshToken, error) {
			return db.RefreshToken{}, pgx.ErrNoRows
		},
	}
	svc := newService(repo)

	_, err := svc.Refresh(context.Background(), "some-refresh-token")
	if !errors.Is(err, ErrRefreshTokenNotFound) {
		t.Fatalf("expected ErrRefreshTokenNotFound, got %v", err)
	}
}

func TestRefreshTokenExpired(t *testing.T) {
	repo := &mockRepo{
		getRefreshToken: func(ctx context.Context, hashToken string) (db.RefreshToken, error) {
			return expiredRefreshToken(), nil
		},
	}
	svc := newService(repo)

	_, err := svc.Refresh(context.Background(), "some-refresh-token")
	if !errors.Is(err, ErrRefreshTokenExpired) {
		t.Fatalf("expected ErrRefreshTokenExpired, got %v", err)
	}
}

func validRefreshToken() db.RefreshToken {
	var exp pgtype.Date
	exp.Scan(time.Now().Add(7 * 24 * time.Hour))

	return db.RefreshToken{ID: "rt-1", UserID: "user-1", HashToken: "hash", ExpiaredAt: exp}
}

func expiredRefreshToken() db.RefreshToken {
	var exp pgtype.Date
	exp.Scan(time.Now().Add(-time.Hour))

	return db.RefreshToken{ID: "rt-1", UserID: "user-1", HashToken: "hash", ExpiaredAt: exp}
}
