package nuvempago

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
)

func TestPersistFutureReceivablesRequiresValidIdentity(t *testing.T) {
	_, err := PersistFutureReceivables(context.Background(), nil, pgtype.UUID{}, pgtype.UUID{}, FutureReport{})
	if err != ErrInvalidFuturePersistenceIdentity {
		t.Fatalf("err=%v, want %v", err, ErrInvalidFuturePersistenceIdentity)
	}
}

func TestFutureImportServiceRequiresDatabase(t *testing.T) {
	if _, err := (*FutureImportService)(nil).Import(context.Background(), t.TempDir()); err != ErrFutureDatabaseUnavailable {
		t.Fatalf("err=%v, want %v", err, ErrFutureDatabaseUnavailable)
	}
}
