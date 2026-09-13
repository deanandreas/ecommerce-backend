package database

import (
	"context"
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5"
)

func RollBack(tx pgx.Tx, ctx context.Context) {
	if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
		slog.Error("failed to rollback", "error", err)
	}
}
