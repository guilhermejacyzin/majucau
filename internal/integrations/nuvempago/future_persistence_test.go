package nuvempago

import (
	"context"
	"testing"
)

func TestFutureImportServiceRequiresDatabase(t *testing.T) {
	if _, err := (*FutureImportService)(nil).Import(context.Background(), t.TempDir()); err != ErrFutureDatabaseUnavailable {
		t.Fatalf("err=%v, want %v", err, ErrFutureDatabaseUnavailable)
	}
}
