package dashboard

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

type fakeRow struct {
	values []any
	err    error
}

func (r fakeRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	if len(dest) != len(r.values) {
		return errors.New("unexpected scan width")
	}
	for i, value := range r.values {
		switch target := dest[i].(type) {
		case *string:
			*target = value.(string)
		case *int64:
			*target = value.(int64)
		default:
			return errors.New("unsupported scan target")
		}
	}
	return nil
}

type fakeDB struct {
	row  fakeRow
	args []any
}

func (d *fakeDB) QueryRow(_ context.Context, _ string, args ...any) pgx.Row {
	d.args = args
	return d.row
}

// Compile-time assertion keeps the test double aligned with the pgx Row
// contract used by Reader without requiring a real database for unit tests.
var _ interface {
	QueryRow(context.Context, string, ...any) pgx.Row
} = (*fakeDB)(nil)

func TestReadDashboardSnapshotUsesNormalizedRowsAndExplicitStates(t *testing.T) {
	db := &fakeDB{row: fakeRow{values: []any{
		"100.0000", int64(2), "40.0000", int64(1),
		"25.0000", int64(1), int64(3), "5.0000", int64(1),
		"80.0000", int64(2), "10.0000", int64(1), "4.0000", int64(1),
		"30.0000", int64(1), int64(2),
	}}}
	reader := &Reader{db: db, now: func() time.Time { return time.Date(2026, 9, 21, 12, 0, 0, 0, time.FixedZone("BRT", -3*60*60)) }}

	snapshot, err := reader.ReadDashboardSnapshot(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.DataState != "PARTIAL" || snapshot.AsOf.Location() != time.UTC {
		t.Fatalf("unexpected snapshot metadata: %+v", snapshot)
	}
	if snapshot.Receivables.Value != "100.0000" || snapshot.Receivables.SourceSystem != "BLING + NUVEM_PAGO" || snapshot.Receivables.Count != 2 {
		t.Fatalf("unexpected receivables metric: %+v", snapshot.Receivables)
	}
	if snapshot.FutureB2C.State != "PROJECTED" || snapshot.ReceiptsMonth.State != "CONFIRMED" || snapshot.PayablesOverdue.Value != "4.0000" {
		t.Fatalf("unexpected source states: %+v", snapshot)
	}
	if len(db.args) != 3 || !reflect.DeepEqual(db.args[0], snapshot.AsOf) {
		t.Fatalf("expected as-of/date query arguments, got %#v", db.args)
	}
}

func TestReadDashboardSnapshotFailsClosedWithoutDatabase(t *testing.T) {
	_, err := (*Reader)(nil).ReadDashboardSnapshot(context.Background())
	if !errors.Is(err, ErrDatabaseUnavailable) {
		t.Fatalf("expected database unavailable, got %v", err)
	}
}

