package bling

import (
	"errors"
	"testing"
)

func TestRunIndependentBlingSyncStepsAttemptsRemainingSourcesAfterFailure(t *testing.T) {
	wantFailure := errors.New("source unavailable")
	var result BlingRawSyncResult
	attempted := make([]string, 0, 3)
	steps := buildBlingSyncSteps(&result, blingDateFilterDue, true,
		func() (APISyncResult, error) {
			attempted = append(attempted, "cash_movements")
			return APISyncResult{Status: "SUCCESS", RecordsRead: 12}, nil
		},
		func() (APISyncResult, error) {
			attempted = append(attempted, "receivables")
			return APISyncResult{Status: "SUCCESS"}, nil
		},
		func() (APISyncResult, error) {
			attempted = append(attempted, "payables")
			return APISyncResult{Status: "FAILED"}, wantFailure
		},
	)
	failures := runIndependentBlingSyncSteps(steps...)

	if len(attempted) != 3 || attempted[0] != "cash_movements" || attempted[1] != "receivables" || attempted[2] != "payables" {
		t.Fatalf("attempted sources = %v, want all three in order", attempted)
	}
	if len(failures) != 1 || !errors.Is(failures[0], wantFailure) {
		t.Fatalf("failures = %v, want the payable error", failures)
	}
	if result.Receivables.Status != "SUCCESS" || result.Payables.Status != "FAILED" || result.CashMovements.Status != "SUCCESS" || result.CashMovements.RecordsRead != 12 {
		t.Fatalf("results were not retained independently: receivables=%+v payables=%+v cash=%+v", result.Receivables, result.Payables, result.CashMovements)
	}

	result, err := finishBlingRawSyncResult(result, failures)
	if !errors.Is(err, wantFailure) || result.Status != "PARTIAL" {
		t.Fatalf("finished result = %+v, error = %v; want PARTIAL and original failure", result, err)
	}
}

func TestBuildBlingSyncStepsKeepsManualDateFilterScope(t *testing.T) {
	var result BlingRawSyncResult
	var attempted []string
	makeRun := func(name string) func() (APISyncResult, error) {
		return func() (APISyncResult, error) {
			attempted = append(attempted, name)
			return APISyncResult{Status: "SUCCESS"}, nil
		}
	}

	steps := buildBlingSyncSteps(&result, blingDateFilterReceived, false, nil, makeRun("receivables"), makeRun("payables"))
	if len(steps) != 1 {
		t.Fatalf("received-date sync has %d steps, want only receivables", len(steps))
	}
	if failures := runIndependentBlingSyncSteps(steps...); len(failures) != 0 {
		t.Fatalf("received-date sync failures = %v", failures)
	}
	if len(attempted) != 1 || attempted[0] != "receivables" {
		t.Fatalf("received-date sources = %v, want only receivables", attempted)
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
