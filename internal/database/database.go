package database

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	db "github.com/deanandreas/ecommerce-api/internal/database/sqlc"
	"github.com/jackc/pgx/v5/pgxpool"
)

type pSQL struct {
	*pgxpool.Pool
	*db.Queries
}

var (
	dbInstance *pSQL
	mu         sync.Mutex
)

// GetDB now accepts a context, which is standard practice for pgx operations.
func GetDB(ctx context.Context, dbURL string) (*pSQL, error) {
	if dbInstance != nil {
		return dbInstance, nil
	}

	mu.Lock()
	defer mu.Unlock()

	if dbInstance != nil {
		return dbInstance, nil
	}

	var err error
	dbInstance, err = dbConnection(ctx, dbURL)
	if err != nil {
		return nil, err
	}

	slog.Info("connected to database successfully")
	return dbInstance, nil
}

func dbConnection(ctx context.Context, dbURL string) (*pSQL, error) {
	if dbURL == "" {
		dbURL = os.Getenv("DATABASE_URL")
	}

	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %v", err)
	}

	config.MaxConns = 200
	config.MinConns = 0
	config.MaxConnIdleTime = 30 * time.Second
	config.MaxConnLifetime = 20 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %v", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("database connection is not alive: %v", err)
	}

	return &pSQL{Pool: pool, Queries: db.New(pool)}, nil
}

func (p *pSQL) Close() {
	slog.Info("database closing...")
	p.Pool.Close()
}

func (p *pSQL) Health(ctx context.Context) map[string]string {
	stats := make(map[string]string)

	if err := p.Ping(ctx); err != nil {
		stats["status"] = "down"
		stats["message"] = fmt.Sprintf("database is not alive %v", err)
		stats["is_healthy"] = "false"
		return stats
	}

	stats["status"] = "up"
	stats["is_healthy"] = "true"

	return stats
}
