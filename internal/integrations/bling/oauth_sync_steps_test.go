package bling

import (
	"errors"
	"testing"
)

func TestRunIndependentBlingSyncStepsAttemptsRemainingSourcesAfterFailure(t *testing.T) {
	wantFailure := errors.New("source unavailable")
	var receivables, payables, cashMovements APISyncResult
	attempted := make([]string, 0, 3)
	failures := runIndependentBlingSyncSteps(
		blingSyncStep{target: &receivables, run: func() (APISyncResult, error) {
			attempted = append(attempted, "receivables")
			return APISyncResult{Status: "SUCCESS"}, nil
		}},
		blingSyncStep{target: &payables, run: func() (APISyncResult, error) {
			attempted = append(attempted, "payables")
			return APISyncResult{Status: "FAILED"}, wantFailure
		}},
		blingSyncStep{target: &cashMovements, run: func() (APISyncResult, error) {
			attempted = append(attempted, "cash_movements")
			return APISyncResult{Status: "SUCCESS", RecordsRead: 12}, nil
		}},
	)

	if len(attempted) != 3 || attempted[0] != "receivables" || attempted[1] != "payables" || attempted[2] != "cash_movements" {
		t.Fatalf("attempted sources = %v, want all three in order", attempted)
	}
	if len(failures) != 1 || !errors.Is(failures[0], wantFailure) {
		t.Fatalf("failures = %v, want the payable error", failures)
	}
	if receivables.Status != "SUCCESS" || payables.Status != "FAILED" || cashMovements.Status != "SUCCESS" || cashMovements.RecordsRead != 12 {
		t.Fatalf("results were not retained independently: receivables=%+v payables=%+v cash=%+v", receivables, payables, cashMovements)
	}

	result, err := finishBlingRawSyncResult(BlingRawSyncResult{Receivables: receivables, Payables: payables, CashMovements: cashMovements}, failures)
	if !errors.Is(err, wantFailure) || result.Status != "PARTIAL" {
		t.Fatalf("finished result = %+v, error = %v; want PARTIAL and original failure", result, err)
	}
}

func TestFinishBlingRawSyncResultDistinguishesTotalFailureAndPartialSource(t *testing.T) {
	failed, err := finishBlingRawSyncResult(BlingRawSyncResult{Payables: APISyncResult{Status: "FAILED"}}, []error{errors.New("failed")})
	if err == nil || failed.Status != "FAILED" {
		t.Fatalf("total failure result = %+v, error = %v, want FAILED", failed, err)
	}

	partial, err := finishBlingRawSyncResult(BlingRawSyncResult{Payables: APISyncResult{Status: "PARTIAL"}}, nil)
	if err != nil || partial.Status != "PARTIAL" {
		t.Fatalf("partial source result = %+v, error = %v, want PARTIAL", partial, err)
	}

	success, err := finishBlingRawSyncResult(BlingRawSyncResult{CashMovements: APISyncResult{Status: "SUCCESS"}}, nil)
	if err != nil || success.Status != "SUCCESS" {
		t.Fatalf("success result = %+v, error = %v, want SUCCESS", success, err)
	}
}
