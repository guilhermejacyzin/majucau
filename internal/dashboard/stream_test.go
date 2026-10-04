package dashboard

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"reflect"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"majucau.local/financial-intelligence/internal/application"
)

type scriptedDashboardDB struct {
	tx      pgx.Tx
	options pgx.TxOptions
}

func (db *scriptedDashboardDB) BeginTx(_ context.Context, options pgx.TxOptions) (pgx.Tx, error) {
	db.options = options
	return db.tx, nil
}

type scriptedDashboardTx struct {
	pgx.Tx
	summary     pgx.Row
	summaryArgs []any
	queryArgs   [][]any
	queryRows   []pgx.Rows
	committed   bool
}

func (tx *scriptedDashboardTx) QueryRow(_ context.Context, _ string, args ...any) pgx.Row {
	tx.summaryArgs = append([]any(nil), args...)
	return tx.summary
}

func (tx *scriptedDashboardTx) Query(_ context.Context, _ string, args ...any) (pgx.Rows, error) {
	tx.queryArgs = append(tx.queryArgs, append([]any(nil), args...))
	if len(tx.queryRows) == 0 {
		return nil, pgx.ErrNoRows
	}
	rows := tx.queryRows[0]
	tx.queryRows = tx.queryRows[1:]
	return rows, nil
}

func (tx *scriptedDashboardTx) Commit(context.Context) error {
	tx.committed = true
	return nil
}

func (*scriptedDashboardTx) Rollback(context.Context) error { return nil }

type scriptedDashboardSummaryRow struct{ values []any }

func (row scriptedDashboardSummaryRow) Scan(dest ...any) error {
	if len(dest) != len(row.values) {
		return pgx.ErrNoRows
	}
	for index, value := range row.values {
		switch target := dest[index].(type) {
		case *string:
			converted, ok := value.(string)
			if !ok {
				return pgx.ErrNoRows
			}
			*target = converted
		case *int64:
			converted, ok := value.(int64)
			if !ok {
				return pgx.ErrNoRows
			}
			*target = converted
		default:
			return pgx.ErrNoRows
		}
	}
	return nil
}

type dashboardRowValues struct {
	fields    [6]string
	oversized bool
}

type scriptedDashboardRows struct {
	pgx.Rows
	values []dashboardRowValues
	index  int
}

func (rows *scriptedDashboardRows) Next() bool {
	if rows.index >= len(rows.values) {
		return false
	}
	rows.index++
	return true
}

func (rows *scriptedDashboardRows) Scan(dest ...any) error {
	if rows.index == 0 || rows.index > len(rows.values) || len(dest) != 7 {
		return pgx.ErrNoRows
	}
	value := rows.values[rows.index-1]
	for index, field := range value.fields {
		target, ok := dest[index].(*string)
		if !ok {
			return pgx.ErrNoRows
		}
		*target = field
	}
	oversized, ok := dest[6].(*bool)
	if !ok {
		return pgx.ErrNoRows
	}
	*oversized = value.oversized
	return nil
}

func (*scriptedDashboardRows) Close()     {}
func (*scriptedDashboardRows) Err() error { return nil }

func TestWriteDashboardSnapshotStreamsRowsInReadOnlyTransaction(t *testing.T) {
	asOf := time.Date(2026, 9, 21, 12, 0, 0, 0, time.UTC)
	tx := &scriptedDashboardTx{
		summary: scriptedDashboardSummaryRow{values: dashboardSummaryValues()},
		queryRows: []pgx.Rows{
			&scriptedDashboardRows{values: []dashboardRowValues{{fields: [6]string{"Cliente", "Nuvem Pago", "2026-09-22", "40.0000", "39.0000", "PROJECTED"}}}},
			&scriptedDashboardRows{values: []dashboardRowValues{{fields: [6]string{"Fornecedor", "NF-1", "2026-09-23", "80.0000", "Insumos", "OPEN"}}}},
		},
	}
	db := &scriptedDashboardDB{tx: tx}
	reader := &Reader{db: db, now: func() time.Time { return asOf }}
	var output bytes.Buffer

	if err := reader.WriteDashboardSnapshot(context.Background(), &output); err != nil {
		t.Fatalf("WriteDashboardSnapshot() error = %v", err)
	}
	var snapshot application.DashboardSnapshot
	if err := json.Unmarshal(output.Bytes(), &snapshot); err != nil {
		t.Fatalf("stream output is not valid snapshot JSON: %v", err)
	}
	if !snapshot.AsOf.Equal(asOf) || snapshot.DataState != "PARTIAL" || snapshot.Receivables.Value != "100.0000" {
		t.Fatalf("unexpected snapshot summary: %+v", snapshot)
	}
	if len(snapshot.ReceivableRows) != 1 || snapshot.ReceivableRows[0].Customer != "Cliente" || len(snapshot.PayableRows) != 1 || snapshot.PayableRows[0].Supplier != "Fornecedor" {
		t.Fatalf("detail rows were not streamed into the snapshot: %+v", snapshot)
	}
	wantOptions := pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly}
	if db.options != wantOptions || !tx.committed {
		t.Fatalf("transaction options/commit = %#v/%t, want %#v/true", db.options, tx.committed, wantOptions)
	}
	wantDates := []any{"2026-09-21", "2026-09-01", "2026-10-01"}
	if !reflect.DeepEqual(tx.summaryArgs, wantDates) {
		t.Fatalf("summary date arguments = %#v, want %#v", tx.summaryArgs, wantDates)
	}
	wantLimits := []any{dashboardFieldMaxBytes, dashboardRecordMaxBytes}
	if len(tx.queryArgs) != 2 || !reflect.DeepEqual(tx.queryArgs[0], wantLimits) || !reflect.DeepEqual(tx.queryArgs[1], wantLimits) {
		t.Fatalf("detail query limits = %#v, want two copies of %#v", tx.queryArgs, wantLimits)
	}
}

func TestReadStreamSummaryUsesBusinessDateAtUtcMonthBoundary(t *testing.T) {
	cases := []struct {
		instant   time.Time
		date      string
		monthFrom string
		monthTo   string
	}{
		{instant: time.Date(2026, 3, 1, 2, 59, 0, 0, time.UTC), date: "2026-02-28", monthFrom: "2026-02-01", monthTo: "2026-03-01"},
		{instant: time.Date(2026, 3, 1, 3, 0, 0, 0, time.UTC), date: "2026-03-01", monthFrom: "2026-03-01", monthTo: "2026-04-01"},
	}
	for _, tc := range cases {
		t.Run(tc.date, func(t *testing.T) {
			tx := &scriptedDashboardTx{summary: scriptedDashboardSummaryRow{values: emptyDashboardSummaryValues()}}
			snapshot, err := readStreamSummary(context.Background(), tx, tc.instant)
			if err != nil {
				t.Fatal(err)
			}
			wantDates := []any{tc.date, tc.monthFrom, tc.monthTo}
			if !reflect.DeepEqual(tx.summaryArgs, wantDates) {
				t.Fatalf("date arguments = %#v, want %#v", tx.summaryArgs, wantDates)
			}
			if !snapshot.AsOf.Equal(tc.instant) || snapshot.AsOf.Location() != time.UTC {
				t.Fatalf("as_of = %s (%s), want %s UTC", snapshot.AsOf, snapshot.AsOf.Location(), tc.instant)
			}
		})
	}
}

func TestDashboardDetailStreamsUseExplicitBounds(t *testing.T) {
	cases := []struct {
		name       string
		stream     func(context.Context, pgx.Tx, io.Writer) error
		fields     [6]string
		firstKey   string
		firstValue string
	}{
		{name: "receivables", stream: streamReceivableRows, fields: [6]string{"Cliente", "Bling", "2026-10-04", "100.00", "100.00", "OPEN"}, firstKey: "customer", firstValue: "Cliente"},
		{name: "payables", stream: streamPayableRows, fields: [6]string{"Fornecedor", "NF-1", "2026-10-04", "100.00", "Insumos", "OPEN"}, firstKey: "supplier", firstValue: "Fornecedor"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rows := &scriptedDashboardRows{values: []dashboardRowValues{{fields: tc.fields}}}
			tx := &scriptedDashboardTx{queryRows: []pgx.Rows{rows}}
			var output bytes.Buffer
			if err := tc.stream(context.Background(), tx, &output); err != nil {
				t.Fatalf("stream() error = %v", err)
			}
			wantLimits := []any{dashboardFieldMaxBytes, dashboardRecordMaxBytes}
			if len(tx.queryArgs) != 1 || !reflect.DeepEqual(tx.queryArgs[0], wantLimits) {
				t.Fatalf("query limits = %#v, want %#v", tx.queryArgs, wantLimits)
			}
			var records []map[string]string
			if err := json.Unmarshal([]byte("["+output.String()+"]"), &records); err != nil {
				t.Fatalf("stream output is not valid JSON: %v", err)
			}
			if len(records) != 1 || records[0][tc.firstKey] != tc.firstValue {
				t.Fatalf("unexpected streamed records: %#v", records)
			}
		})
	}
}

func TestDashboardDetailStreamsRejectOversizedRecordsBeforeWritingThem(t *testing.T) {
	for _, tc := range []struct {
		name   string
		stream func(context.Context, pgx.Tx, io.Writer) error
	}{
		{name: "receivables", stream: streamReceivableRows},
		{name: "payables", stream: streamPayableRows},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rows := &scriptedDashboardRows{values: []dashboardRowValues{{oversized: true}}}
			tx := &scriptedDashboardTx{queryRows: []pgx.Rows{rows}}
			var output bytes.Buffer
			err := tc.stream(context.Background(), tx, &output)
			if err != ErrSnapshotRecordTooLarge {
				t.Fatalf("stream() error = %v, want %v", err, ErrSnapshotRecordTooLarge)
			}
			if output.Len() != 0 {
				t.Fatalf("oversized row was written: %q", output.String())
			}
		})
	}
}

func TestWriteDashboardSnapshotFailsClosedWithoutDatabase(t *testing.T) {
	if err := NewReader(nil).WriteDashboardSnapshot(context.Background(), io.Discard); err != ErrDatabaseUnavailable {
		t.Fatalf("WriteDashboardSnapshot() error = %v, want %v", err, ErrDatabaseUnavailable)
	}
}

func dashboardSummaryValues() []any {
	return []any{
		"100.0000", int64(2), "40.0000", int64(1),
		"25.0000", int64(1), int64(3), "5.0000", int64(1),
		"80.0000", int64(2), "10.0000", int64(1), "4.0000", int64(1),
		"30.0000", int64(1), int64(2),
	}
}

func emptyDashboardSummaryValues() []any {
	return []any{
		"0.0000", int64(0), "0.0000", int64(0),
		"0.0000", int64(0), int64(0), "0.0000", int64(0),
		"0.0000", int64(0), "0.0000", int64(0), "0.0000", int64(0),
		"0.0000", int64(0), int64(0),
	}
}
