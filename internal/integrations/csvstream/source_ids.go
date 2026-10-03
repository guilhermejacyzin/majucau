package csvstream

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/csv"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// RecoverableRowError recognizes CSV errors that encoding/csv reports only
// after consuming the complete physical record. ErrQuote is deliberately
// excluded because it also represents an unterminated quoted field at EOF,
// where later records may have been swallowed. I/O and ambiguous quote errors
// abort the caller's transaction instead of producing an incomplete import.
func RecoverableRowError(err error) bool {
	var parseErr *csv.ParseError
	if !errors.As(err, &parseErr) {
		return false
	}
	return errors.Is(parseErr.Err, csv.ErrBareQuote) || errors.Is(parseErr.Err, csv.ErrFieldCount)
}

// SourceIDSet keeps import deduplication state inside the caller's database
// transaction. Only keyed fingerprints are stored; the source identifiers and
// the per-import key stay in process memory. PostgreSQL drops the temporary
// table on commit or rollback.
type SourceIDSet struct {
	tx  pgx.Tx
	key [32]byte
}

// SourceIDIndex provides exact, bounded-heap deduplication and an ordered
// stream of CSV filenames. Implementations may use a database transaction or
// a private, short-lived local index for previews.
type SourceIDIndex interface {
	CheckAndAdd(ctx context.Context, sourceID, file string, line int) (firstFile string, firstLine int64, duplicate bool, err error)
	QueueCSVFiles(ctx context.Context, folder string) (CSVFileQueue, error)
}

// CSVFileQueue yields ordered CSV filenames without materializing a folder's
// complete directory listing in memory.
type CSVFileQueue interface {
	Next() (string, bool, error)
	IgnoredCount() int
	Close()
}

func NewSourceIDSet(ctx context.Context, tx pgx.Tx) (*SourceIDSet, error) {
	if tx == nil {
		return nil, errors.New("CSV source ID transaction is required")
	}
	set := &SourceIDSet{tx: tx}
	if _, err := rand.Read(set.key[:]); err != nil {
		return nil, fmt.Errorf("generate CSV deduplication key: %w", err)
	}
	if _, err := tx.Exec(ctx, `
CREATE TEMP TABLE csv_import_seen_source_ids (
  fingerprint bytea PRIMARY KEY,
  first_file text NOT NULL,
  first_line bigint NOT NULL
) ON COMMIT DROP
`); err != nil {
		set.Close()
		return nil, fmt.Errorf("create temporary CSV deduplication table: %w", err)
	}
	return set, nil
}

// CheckAndAdd returns duplicate=true and the first location when sourceID was
// already seen in this import. The database stores an HMAC fingerprint rather
// than the raw identifier, and its row count does not grow Go heap usage.
func (s *SourceIDSet) CheckAndAdd(ctx context.Context, sourceID, file string, line int) (firstFile string, firstLine int64, duplicate bool, err error) {
	if s == nil || s.tx == nil || sourceID == "" {
		return "", 0, false, errors.New("CSV source ID set is not initialized")
	}
	mac := hmac.New(sha256.New, s.key[:])
	_, _ = mac.Write([]byte(sourceID))
	fingerprint := mac.Sum(nil)
	err = s.tx.QueryRow(ctx, `
INSERT INTO pg_temp.csv_import_seen_source_ids (fingerprint, first_file, first_line)
VALUES ($1, $2, $3)
ON CONFLICT (fingerprint) DO NOTHING
RETURNING first_file, first_line
`, fingerprint, file, int64(line)).Scan(&firstFile, &firstLine)
	if err == nil {
		return "", 0, false, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return "", 0, false, fmt.Errorf("record CSV source ID fingerprint: %w", err)
	}
	err = s.tx.QueryRow(ctx, `
SELECT first_file, first_line
FROM pg_temp.csv_import_seen_source_ids
WHERE fingerprint = $1
`, fingerprint).Scan(&firstFile, &firstLine)
	if err != nil {
		return "", 0, false, fmt.Errorf("read first CSV source ID location: %w", err)
	}
	return firstFile, firstLine, true, nil
}

func (s *SourceIDSet) Close() {
	if s == nil {
		return
	}
	for i := range s.key {
		s.key[i] = 0
	}
	s.tx = nil
}
