package dbtx

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

const rollbackTimeout = 5 * time.Second

type Rollbacker interface {
	Rollback(context.Context) error
}

// Rollback attempts transaction cleanup even when the operation context has
// been cancelled. A committed or already rolled back pgx transaction is a
// normal terminal state; other cleanup failures are returned to the caller.
func Rollback(ctx context.Context, tx Rollbacker) error {
	if ctx == nil {
		return errors.New("transaction rollback context is required")
	}
	if tx == nil {
		return errors.New("transaction rollback target is required")
	}

	rollbackCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), rollbackTimeout)
	defer cancel()
	if err := tx.Rollback(rollbackCtx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
		return fmt.Errorf("database transaction rollback failed: %w", err)
	}
	return nil
}
