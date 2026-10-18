package main

import (
	"context"
	"errors"
	"testing"

	"majucau.local/financial-intelligence/internal/integrations/bling"
)

type fakeDailyBlingSyncer struct {
	filter bling.ReceivablesFilter
	result bling.BlingRawSyncResult
	err    error
	calls  int
}

func (f *fakeDailyBlingSyncer) Sync(_ context.Context, filter bling.ReceivablesFilter) (bling.BlingRawSyncResult, error) {
	f.calls++
	f.filter = filter
	return f.result, f.err
}

func TestRunDailyBlingSyncRelistsAllAccountsWithoutDateWindow(t *testing.T) {
	syncer := &fakeDailyBlingSyncer{result: bling.BlingRawSyncResult{Status: "SUCCESS"}}
	result, err := runDailyBlingSync(context.Background(), syncer)
	if err != nil || result.Status != "SUCCESS" || syncer.calls != 1 {
		t.Fatalf("daily sync result=%+v error=%v calls=%d", result, err, syncer.calls)
	}
	if syncer.filter.Page != 0 || syncer.filter.Limit != 0 || syncer.filter.DueDateFrom != "" || syncer.filter.DueDateTo != "" || syncer.filter.ReceivedDateFrom != "" || syncer.filter.ReceivedDateTo != "" || syncer.filter.PaymentDateFrom != "" || syncer.filter.PaymentDateTo != "" {
		t.Fatalf("daily run must request all pages and all dates, got filter=%+v", syncer.filter)
	}
}

func TestRunDailyBlingSyncRejectsMissingOrIncompleteSource(t *testing.T) {
	if _, err := runDailyBlingSync(context.Background(), nil); !errors.Is(err, ErrDailyBlingUnavailable) {
		t.Fatalf("missing source error = %v", err)
	}
	syncer := &fakeDailyBlingSyncer{result: bling.BlingRawSyncResult{Status: "PARTIAL"}}
	if _, err := runDailyBlingSync(context.Background(), syncer); !errors.Is(err, ErrDailyBlingIncomplete) {
		t.Fatalf("incomplete result error = %v", err)
	}
}

func TestDailyBlingSyncErrorCodeDoesNotExposeUnderlyingError(t *testing.T) {
	err := errors.New("secret token and account data")
	if got := dailyBlingSyncErrorCode(err); got != "BLING_SYNC_FAILED" {
		t.Fatalf("error code = %q", got)
	}
}
