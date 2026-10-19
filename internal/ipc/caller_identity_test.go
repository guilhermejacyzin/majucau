package ipc

import (
	"context"
	"testing"
)

func TestCallerWindowsSIDCanOnlyBeReadFromValidatedIPCContext(t *testing.T) {
	if _, ok := CallerWindowsSID(context.Background()); ok {
		t.Fatal("ordinary context unexpectedly contained a caller SID")
	}
	ctx := WithCallerWindowsSID(context.Background(), "S-1-5-21-1-2-3-1001")
	if sid, ok := CallerWindowsSID(ctx); !ok || sid != "S-1-5-21-1-2-3-1001" {
		t.Fatalf("caller SID = %q, ok=%v", sid, ok)
	}
	invalid := WithCallerWindowsSID(context.Background(), "S-1-5-21-not-a-number")
	if _, ok := CallerWindowsSID(invalid); ok {
		t.Fatal("invalid SID was accepted")
	}
}
