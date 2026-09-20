package bling

import (
	"context"
	"errors"
	"testing"
)

func TestReceiptImportServiceFailsClosedWithoutPool(t *testing.T) {
	var service *ReceiptImportService
	_, err := service.Import(context.Background(), "C:\\imports")
	if !errors.Is(err, ErrReceiptDatabaseUnavailable) {
		t.Fatalf("missing pool must be explicit: %v", err)
	}
}
