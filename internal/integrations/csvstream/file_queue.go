package csvstream

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/jackc/pgx/v5"
)

const directoryReadBatchSize = 256
const temporarySortMemory = "16MB"

// FileQueue spills CSV filenames to a transaction-local table, so folder size
// does not grow the worker's heap. PostgreSQL orders the names before they are
// read back, preserving deterministic import order without a large Go slice.
type FileQueue struct {
	tx           pgx.Tx
	ctx          context.Context
	batch        []queuedFile
	batchIndex   int
	lastSortKey  string
	lastName     string
	started      bool
	done         bool
	ignoredCount int
}

type queuedFile struct {
	sortKey string
	name    string
}

func (s *SourceIDSet) QueueCSVFiles(ctx context.Context, folder string) (CSVFileQueue, error) {
	if s == nil || s.tx == nil {
		return nil, errors.New("CSV source ID set is not initialized")
	}
	directory, err := os.Open(folder)
	if err != nil {
		return nil, fmt.Errorf("open CSV import folder: %w", err)
	}
	defer directory.Close()

	if _, err := s.tx.Exec(ctx, `
CREATE TEMP TABLE csv_import_files (
	  sort_key text COLLATE "C" NOT NULL,
	  name text COLLATE "C" PRIMARY KEY
) ON COMMIT DROP
`); err != nil {
		return nil, fmt.Errorf("create temporary CSV file queue: %w", err)
	}
	if _, err := s.tx.Exec(ctx, fmt.Sprintf(`SET LOCAL work_mem = '%s'`, temporarySortMemory)); err != nil {
		return nil, fmt.Errorf("bound temporary CSV sort memory: %w", err)
	}
	if _, err := s.tx.Exec(ctx, fmt.Sprintf(`SET LOCAL maintenance_work_mem = '%s'`, temporarySortMemory)); err != nil {
		return nil, fmt.Errorf("bound temporary CSV index memory: %w", err)
	}

	queue := &FileQueue{}
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		entries, readErr := directory.ReadDir(directoryReadBatchSize)
		names := make([]string, 0, len(entries))
		for _, entry := range entries {
			info, err := entry.Info()
			if err != nil {
				return nil, fmt.Errorf("inspect import folder entry: %w", err)
			}
			if !info.Mode().IsRegular() || !strings.EqualFold(filepath.Ext(entry.Name()), ".csv") {
				queue.ignoredCount++
				continue
			}
			names = append(names, entry.Name())
		}
		if len(names) > 0 {
			if _, err := s.tx.CopyFrom(ctx, pgx.Identifier{"pg_temp", "csv_import_files"}, []string{"sort_key", "name"}, pgx.CopyFromSlice(len(names), func(index int) ([]any, error) {
				return []any{strings.ToLower(names[index]), names[index]}, nil
			})); err != nil {
				return nil, fmt.Errorf("queue CSV filenames: %w", err)
			}
		}
		if errors.Is(readErr, io.EOF) {
			break
		}
		if readErr != nil {
			return nil, fmt.Errorf("read import folder entries: %w", readErr)
		}
	}
	if err := directory.Close(); err != nil {
		return nil, fmt.Errorf("close import folder: %w", err)
	}
	if _, err := s.tx.Exec(ctx, `CREATE INDEX csv_import_files_order_idx ON pg_temp.csv_import_files (sort_key COLLATE "C", name COLLATE "C")`); err != nil {
		return nil, fmt.Errorf("index temporary CSV filenames: %w", err)
	}
	queue.tx = s.tx
	queue.ctx = ctx
	return queue, nil
}

func (q *FileQueue) Next() (string, bool, error) {
	if q == nil || q.tx == nil {
		return "", false, errors.New("CSV file queue is not initialized")
	}
	if q.batchIndex >= len(q.batch) {
		if err := q.loadBatch(); err != nil {
			return "", false, err
		}
		if len(q.batch) == 0 {
			return "", false, nil
		}
	}
	name := q.batch[q.batchIndex].name
	q.batchIndex++
	return name, true, nil
}

// loadBatch fully consumes and closes the PostgreSQL result before returning
// filenames to the caller. The caller can then use the same transaction to
// deduplicate and persist CSV rows without pgx reporting a busy connection.
func (q *FileQueue) loadBatch() error {
	if q.done {
		return nil
	}
	if err := q.ctx.Err(); err != nil {
		return err
	}
	var rows pgx.Rows
	var err error
	if !q.started {
		rows, err = q.tx.Query(q.ctx, `
SELECT sort_key, name
FROM pg_temp.csv_import_files
ORDER BY sort_key COLLATE "C", name COLLATE "C"
LIMIT $1
`, directoryReadBatchSize)
		q.started = true
	} else {
		rows, err = q.tx.Query(q.ctx, `
SELECT sort_key, name
FROM pg_temp.csv_import_files
WHERE (sort_key COLLATE "C", name COLLATE "C") > ($1::text COLLATE "C", $2::text COLLATE "C")
ORDER BY sort_key COLLATE "C", name COLLATE "C"
LIMIT $3
`, q.lastSortKey, q.lastName, directoryReadBatchSize)
	}
	if err != nil {
		return fmt.Errorf("read ordered CSV filename batch: %w", err)
	}
	q.batch = make([]queuedFile, 0, directoryReadBatchSize)
	q.batchIndex = 0
	for rows.Next() {
		var file queuedFile
		if err := rows.Scan(&file.sortKey, &file.name); err != nil {
			rows.Close()
			return fmt.Errorf("scan ordered CSV filename: %w", err)
		}
		q.batch = append(q.batch, file)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return fmt.Errorf("read ordered CSV filename batch: %w", err)
	}
	rows.Close()
	if len(q.batch) == 0 {
		q.done = true
		return nil
	}
	last := q.batch[len(q.batch)-1]
	q.lastSortKey = last.sortKey
	q.lastName = last.name
	return nil
}

func (q *FileQueue) IgnoredCount() int {
	if q == nil {
		return 0
	}
	return q.ignoredCount
}

func (q *FileQueue) Close() {
	if q != nil {
		q.batch = nil
		q.tx = nil
		q.ctx = nil
	}
}
