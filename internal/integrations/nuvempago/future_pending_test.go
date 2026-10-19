package nuvempago

import (
	"errors"
	"strings"
	"testing"
)

func TestFutureImportPendingAuditUsesSafeCountsAndFixedMessages(t *testing.T) {
	pending := FutureImportPending{
		NoCompatibleCSV: true, DeferredFiles: 2, StaleFiles: 1, UnsupportedFiles: 3,
		DownloadFailedFiles: 4, InvalidMetadataFiles: 5,
	}
	code, message, count, err := pending.auditDetails()
	if err != nil {
		t.Fatal(err)
	}
	if code != "NUVEM_PAGO_FUTURE_SOURCE_PENDING" || count != 15 {
		t.Fatalf("pending audit code=%q count=%d", code, count)
	}
	for _, expected := range []string{"nenhum CSV compatível", "2 arquivo(s)", "1 arquivo(s) estão antigos", "3 arquivo(s)", "4 arquivo(s)", "5 arquivo(s)"} {
		if !strings.Contains(message, expected) {
			t.Errorf("audit message %q does not contain %q", message, expected)
		}
	}
	for _, forbidden := range []string{"secret.csv", "customer@example.test", "file_123"} {
		if strings.Contains(message, forbidden) {
			t.Errorf("audit message contains source detail %q", forbidden)
		}
	}
}

func TestFutureImportPendingRejectsInvalidCounts(t *testing.T) {
	for _, pending := range []FutureImportPending{
		{DeferredFiles: -1},
		{StaleFiles: -1},
		{UnsupportedFiles: -1},
		{DownloadFailedFiles: -1},
		{InvalidMetadataFiles: -1},
		{DeferredFiles: int64(^uint64(0) >> 1), UnsupportedFiles: 1},
	} {
		if _, _, _, err := pending.auditDetails(); !errors.Is(err, ErrFuturePendingInvalid) {
			t.Errorf("pending summary %#v error = %v", pending, err)
		}
	}
}

func TestFutureImportPendingWithInvalidMetadataIsPartial(t *testing.T) {
	pending := FutureImportPending{InvalidMetadataFiles: 1}
	if !pending.hasPending() {
		t.Fatal("invalid source metadata must mark the import as partial")
	}
	_, message, count, err := pending.auditDetails()
	if err != nil || count != 1 || !strings.Contains(message, "1 arquivo(s) têm informações incompletas") {
		t.Fatalf("pending metadata audit message=%q count=%d error=%v", message, count, err)
	}
}
