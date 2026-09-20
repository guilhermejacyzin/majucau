package bling

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
)

func TestPersistReceiptsRequiresDatabaseIdentitiesBeforeOpeningWork(t *testing.T) {
	_, err := PersistReceipts(context.Background(), nil, pgtype.UUID{}, pgtype.UUID{}, ReceiptsReport{})
	if !errors.Is(err, ErrInvalidPersistenceIdentity) {
		t.Fatalf("invalid identities must fail before DB use: %v", err)
	}
}

func TestPersistReceiptsRequiresTransactionAfterIdentityValidation(t *testing.T) {
	connectionID := pgtype.UUID{Valid: true}
	batchID := pgtype.UUID{Valid: true}
	_, err := PersistReceipts(context.Background(), nil, connectionID, batchID, ReceiptsReport{})
	if err == nil || err.Error() != "receipt persistence transaction is required" {
		t.Fatalf("missing transaction must fail safely: %v", err)
	}
}
