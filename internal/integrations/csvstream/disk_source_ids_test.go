package csvstream

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestDiskSourceIDSetDeduplicatesWithoutPersistingPlaintextAndCleansUp(t *testing.T) {
	ctx := context.Background()
	folder := t.TempDir()
	const sensitiveFilename = "sensitive-preview-filename.csv"
	if err := os.WriteFile(filepath.Join(folder, sensitiveFilename), []byte("source data"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(folder, "ignore.txt"), []byte("ignored"), 0o600); err != nil {
		t.Fatal(err)
	}

	set, err := NewDiskSourceIDSet(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = set.Close() })
	workspace := set.dir
	firstID := "sensitive-source-id-0000"
	lastID := "sensitive-source-id-0749"
	for index := 0; index < 750; index++ {
		sourceID := fmt.Sprintf("sensitive-source-id-%04d", index)
		_, _, duplicate, err := set.CheckAndAdd(ctx, sourceID, sensitiveFilename, index+2)
		if err != nil {
			t.Fatal(err)
		}
		if duplicate {
			t.Fatalf("first insertion of %q was reported as duplicate", sourceID)
		}
	}
	if set.count != 750 || set.capacity < 2048 {
		t.Fatalf("index did not grow on disk as expected: count=%d capacity=%d", set.count, set.capacity)
	}
	file, line, duplicate, err := set.CheckAndAdd(ctx, firstID, "other.csv", 900)
	if err != nil {
		t.Fatal(err)
	}
	if !duplicate || file != sensitiveFilename || line != 2 {
		t.Fatalf("duplicate did not retain its first location: file=%q line=%d duplicate=%v", file, line, duplicate)
	}
	file, line, duplicate, err = set.CheckAndAdd(ctx, lastID, "other.csv", 901)
	if err != nil {
		t.Fatal(err)
	}
	if !duplicate || file != sensitiveFilename || line != 751 {
		t.Fatalf("last duplicate did not retain its first location: file=%q line=%d duplicate=%v", file, line, duplicate)
	}

	queue, err := set.QueueCSVFiles(ctx, folder)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := queue.Close(); err != nil {
			t.Errorf("close temporary CSV file queue: %v", err)
		}
	})
	if queue.IgnoredCount() != 1 {
		t.Fatalf("ignored file count = %d, want 1", queue.IgnoredCount())
	}
	queuedName, ok, err := queue.Next()
	if err != nil || !ok || queuedName != sensitiveFilename {
		t.Fatalf("ordered queue returned %q, %v, %v", queuedName, ok, err)
	}
	if queuedName, ok, err = queue.Next(); err != nil || ok {
		t.Fatalf("queue should be exhausted: %q, %v, %v", queuedName, ok, err)
	}
	if err := queue.Close(); err != nil {
		t.Fatal(err)
	}

	entries, err := os.ReadDir(workspace)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		contents, err := os.ReadFile(filepath.Join(workspace, entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		for _, sensitive := range []string{firstID, lastID, sensitiveFilename} {
			if bytes.Contains(contents, []byte(sensitive)) {
				t.Fatalf("temporary file %q contains sensitive plaintext %q", entry.Name(), sensitive)
			}
		}
	}

	if err := set.Close(); err != nil {
		t.Fatal(err)
	}
	if err := set.Close(); err != nil {
		t.Fatalf("second Close should be idempotent: %v", err)
	}
	if _, err := os.Stat(workspace); !os.IsNotExist(err) {
		t.Fatalf("temporary workspace still exists after Close: %v", err)
	}
	if set.gcm != nil {
		t.Fatal("encryption key state must be released on Close")
	}
}

func TestDiskRunBuilderAddRemovesMergedTemporaryRuns(t *testing.T) {
	ctx := context.Background()
	set, err := NewDiskSourceIDSet(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := set.Close(); err != nil {
			t.Errorf("close temporary CSV source index: %v", err)
		}
	})

	builder := diskRunBuilder{set: set}
	paths := writeTemporaryNameRuns(t, set, diskMergeFanIn)
	for _, path := range paths {
		if err := builder.add(ctx, path); err != nil {
			t.Fatalf("add temporary CSV filename run: %v", err)
		}
	}
	for _, path := range paths {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("merged temporary CSV filename run %q still exists: %v", path, err)
		}
	}

	mergedPath, err := builder.finish(ctx)
	if err != nil {
		t.Fatalf("finish temporary CSV filename runs: %v", err)
	}
	if _, err := os.Stat(mergedPath); err != nil {
		t.Fatalf("final temporary CSV filename run is missing: %v", err)
	}
}

func TestDiskRunBuilderFinishRemovesIntermediateTemporaryRuns(t *testing.T) {
	ctx := context.Background()
	set, err := NewDiskSourceIDSet(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := set.Close(); err != nil {
			t.Errorf("close temporary CSV source index: %v", err)
		}
	})

	paths := writeTemporaryNameRuns(t, set, diskMergeFanIn+1)
	builder := diskRunBuilder{set: set, levels: [][]string{paths}}
	mergedPath, err := builder.finish(ctx)
	if err != nil {
		t.Fatalf("finish temporary CSV filename runs: %v", err)
	}
	for _, path := range paths {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("intermediate temporary CSV filename run %q still exists: %v", path, err)
		}
	}
	if _, err := os.Stat(mergedPath); err != nil {
		t.Fatalf("final temporary CSV filename run is missing: %v", err)
	}
}

func writeTemporaryNameRuns(t *testing.T, set *DiskSourceIDSet, count int) []string {
	t.Helper()
	paths := make([]string, 0, count)
	for index := 0; index < count; index++ {
		path, err := set.writeNameRun([]string{fmt.Sprintf("temporary-%04d.csv", index)})
		if err != nil {
			t.Fatalf("write temporary CSV filename run: %v", err)
		}
		paths = append(paths, path)
	}
	return paths
}
