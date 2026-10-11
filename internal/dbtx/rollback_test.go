package dbtx

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

type rollbackStub struct {
	ctxErr      error
	deadline    time.Time
	hasDeadline bool
	traceValue  any
	err         error
	calls       int
}

type rollbackContextKey struct{}

func (s *rollbackStub) Rollback(ctx context.Context) error {
	s.ctxErr = ctx.Err()
	s.deadline, s.hasDeadline = ctx.Deadline()
	s.traceValue = ctx.Value(rollbackContextKey{})
	s.calls++
	return s.err
}

func TestRollbackUsesDetachedBoundedContextAndIgnoresClosedTransaction(t *testing.T) {
	parent, cancel := context.WithCancel(context.WithValue(context.Background(), rollbackContextKey{}, "trace"))
	cancel()
	tx := &rollbackStub{err: pgx.ErrTxClosed}

	if err := Rollback(parent, tx); err != nil {
		t.Fatalf("Rollback() error = %v, want nil for a closed transaction", err)
	}
	if tx.calls != 1 {
		t.Fatalf("Rollback() called %d times, want 1", tx.calls)
	}
	if tx.ctxErr != nil {
		t.Fatalf("rollback context unexpectedly inherited cancellation: %v", tx.ctxErr)
	}
	if tx.traceValue != "trace" {
		t.Fatal("rollback context did not preserve trace values")
	}
	if !tx.hasDeadline || time.Until(tx.deadline) <= 0 || time.Until(tx.deadline) > rollbackTimeout {
		t.Fatalf("rollback context deadline = %v, want within %s", tx.deadline, rollbackTimeout)
	}
}

func TestRollbackReportsCleanupFailure(t *testing.T) {
	wantErr := errors.New("database connection lost")
	tx := &rollbackStub{err: wantErr}
	if err := Rollback(context.Background(), tx); !errors.Is(err, wantErr) {
		t.Fatalf("Rollback() error = %v, want wrapped %v", err, wantErr)
	}
}

func TestRollbackRejectsMissingInputs(t *testing.T) {
	if err := Rollback(nil, &rollbackStub{}); err == nil {
		t.Fatal("Rollback() accepted a nil context")
	}
	if err := Rollback(context.Background(), nil); err == nil {
		t.Fatal("Rollback() accepted a nil transaction")
	}
}
