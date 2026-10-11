package csvstream

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestPostgresCSVFileQueueStreamsOrderedBatches(t *testing.T) {
	pool, ctx := csvFileQueueTestPool(t)
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin PostgreSQL transaction: %v", err)
	}
	t.Cleanup(func() { _ = tx.Rollback(ctx) })
	set, err := NewSourceIDSet(ctx, tx)
	if err != nil {
		t.Fatalf("create CSV source ID set: %v", err)
	}
	t.Cleanup(set.Close)

	folder := t.TempDir()
	want := make([]string, 0, directoryReadBatchSize*2+3)
	for index := 0; index < directoryReadBatchSize*2+3; index++ {
		name := fmt.Sprintf("file-%04d.csv", index)
		if err := os.WriteFile(filepath.Join(folder, name), nil, 0o600); err != nil {
			t.Fatalf("create CSV fixture %q: %v", name, err)
		}
		want = append(want, name)
	}
	if err := os.WriteFile(filepath.Join(folder, "ignore.txt"), nil, 0o600); err != nil {
		t.Fatalf("create ignored fixture: %v", err)
	}
	if err := os.Mkdir(filepath.Join(folder, "directory.csv"), 0o700); err != nil {
		t.Fatalf("create CSV-named directory fixture: %v", err)
	}
	sort.Slice(want, func(i, j int) bool { return diskNameLess(want[i], want[j]) })

	queue, err := set.QueueCSVFiles(ctx, folder)
	if err != nil {
		t.Fatalf("queue CSV filenames: %v", err)
	}
	t.Cleanup(func() { _ = queue.Close() })
	if got := queue.IgnoredCount(); got != 2 {
		t.Fatalf("ignored entry count = %d, want 2", got)
	}
	for index, expected := range want {
		got, ok, err := queue.Next()
		if err != nil || !ok || got != expected {
			t.Fatalf("queued filename %d = %q, %v, %v; want %q, true, nil", index, got, ok, err, expected)
		}
	}
	if got, ok, err := queue.Next(); err != nil || ok {
		t.Fatalf("queue after all batches = %q, %v, %v; want empty, false, nil", got, ok, err)
	}
}

func TestPostgresCSVFileQueueHonorsCancellationWithBufferedNames(t *testing.T) {
	pool, ctx := csvFileQueueTestPool(t)
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin PostgreSQL transaction: %v", err)
	}
	t.Cleanup(func() { _ = tx.Rollback(ctx) })
	set, err := NewSourceIDSet(ctx, tx)
	if err != nil {
		t.Fatalf("create CSV source ID set: %v", err)
	}
	t.Cleanup(set.Close)
	folder := t.TempDir()
	for index := 0; index < 2; index++ {
		name := fmt.Sprintf("file-%04d.csv", index)
		if err := os.WriteFile(filepath.Join(folder, name), nil, 0o600); err != nil {
			t.Fatalf("create CSV fixture %q: %v", name, err)
		}
	}

	queueCtx, queueCancel := context.WithCancel(ctx)
	t.Cleanup(queueCancel)
	queue, err := set.QueueCSVFiles(queueCtx, folder)
	if err != nil {
		t.Fatalf("queue CSV filenames: %v", err)
	}
	t.Cleanup(func() { _ = queue.Close() })
	if _, ok, err := queue.Next(); err != nil || !ok {
		t.Fatalf("read first queued filename: ok=%v err=%v", ok, err)
	}
	queueCancel()
	if _, ok, err := queue.Next(); err != context.Canceled || ok {
		t.Fatalf("read after cancellation: ok=%v err=%v, want context.Canceled", ok, err)
	}
}

func csvFileQueueTestPool(t *testing.T) (*pgxpool.Pool, context.Context) {
	t.Helper()
	dsn := os.Getenv("MAJUCAU_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set MAJUCAU_TEST_DATABASE_URL to run the PostgreSQL CSV file queue integration test")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("create PostgreSQL pool: %v", err)
	}
	t.Cleanup(pool.Close)
	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("ping PostgreSQL: %v", err)
	}
	return pool, ctx
}
