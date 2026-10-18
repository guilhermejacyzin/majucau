package bling

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	dailyBlingLockClass int32 = 0x4D414A55
	dailyBlingLockKey   int32 = 0x424C494E
)

var (
	ErrBlingDailyScheduleInvalid = errors.New("daily Bling schedule time is invalid")
	ErrBlingDailySyncLock        = errors.New("daily Bling sync lock is unavailable")
	ErrBlingDailySyncLockRelease = errors.New("daily Bling sync lock could not be released")
)

// SyncDaily runs a full Bling refresh for the scheduled São Paulo business
// date. A PostgreSQL session lock prevents parallel worker instances from
// processing the same daily load, and the batch correlation ID lets a restart
// skip a load whose receivables and payables both finished successfully.
func (s *BlingOAuthService) SyncDaily(ctx context.Context, scheduledAt time.Time) (result BlingRawSyncResult, skipped bool, returnErr error) {
	correlationID, err := dailyBlingCorrelationID(scheduledAt)
	if err != nil {
		return BlingRawSyncResult{}, false, err
	}
	if s == nil || s.pool == nil {
		return BlingRawSyncResult{}, false, ErrBlingAPISyncDatabaseUnavailable
	}
	connection, err := s.pool.Acquire(ctx)
	if err != nil {
		return BlingRawSyncResult{}, false, fmt.Errorf("acquire daily Bling sync lock connection: %w", err)
	}
	if _, err := connection.Exec(ctx, `SELECT pg_advisory_lock($1, $2)`, dailyBlingLockClass, dailyBlingLockKey); err != nil {
		closeHijackedConnection(connection)
		return BlingRawSyncResult{}, false, errors.Join(ErrBlingDailySyncLock, err)
	}
	defer func() {
		if err := releaseDailyBlingLock(connection); err != nil {
			returnErr = errors.Join(returnErr, err)
		}
	}()

	correlation := pgUUID(correlationID)
	completed, err := dailyBlingSyncCompleted(ctx, connection, correlation)
	if err != nil {
		return BlingRawSyncResult{}, false, fmt.Errorf("check daily Bling sync completion: %w", err)
	}
	if completed {
		return BlingRawSyncResult{Status: "SUCCESS"}, true, nil
	}
	result, err = s.sync(ctx, ReceivablesFilter{}, &correlationID)
	return result, false, err
}

func dailyBlingSyncCompleted(ctx context.Context, connection *pgxpool.Conn, correlation pgtype.UUID) (bool, error) {
	var completed bool
	err := connection.QueryRow(ctx, `
		SELECT
			EXISTS (SELECT 1 FROM integration_sync_batches WHERE correlation_id = $1 AND resource = $2 AND status = 'SUCCESS')
			AND EXISTS (SELECT 1 FROM integration_sync_batches WHERE correlation_id = $1 AND resource = $3 AND status = 'SUCCESS')`,
		correlation, ReceivablesAPIResource, PayablesAPIResource,
	).Scan(&completed)
	return completed, err
}

func dailyBlingCorrelationID(scheduledAt time.Time) (uuid.UUID, error) {
	if scheduledAt.IsZero() {
		return uuid.Nil, ErrBlingDailyScheduleInvalid
	}
	location, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		return uuid.Nil, errors.Join(ErrBlingDailyScheduleInvalid, err)
	}
	local := scheduledAt.In(location)
	if local.Hour() != 9 || local.Minute() != 30 || local.Second() != 0 || local.Nanosecond() != 0 {
		return uuid.Nil, ErrBlingDailyScheduleInvalid
	}
	return uuid.NewSHA1(uuid.NameSpaceOID, []byte("majucau/bling/daily-sync/v1:"+local.Format("2006-01-02"))), nil
}

func pgUUID(value uuid.UUID) pgtype.UUID {
	var bytes [16]byte
	copy(bytes[:], value[:])
	return pgtype.UUID{Bytes: bytes, Valid: true}
}

func releaseDailyBlingLock(connection *pgxpool.Conn) error {
	unlockCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var unlocked bool
	if err := connection.QueryRow(unlockCtx, `SELECT pg_advisory_unlock($1, $2)`, dailyBlingLockClass, dailyBlingLockKey).Scan(&unlocked); err != nil || !unlocked {
		closeErr := closeHijackedConnection(connection)
		if err == nil {
			err = errors.New("database did not confirm the daily sync lock release")
		}
		return errors.Join(ErrBlingDailySyncLockRelease, err, closeErr)
	}
	connection.Release()
	return nil
}

func closeHijackedConnection(connection *pgxpool.Conn) error {
	if connection == nil {
		return nil
	}
	databaseConnection := connection.Hijack()
	if databaseConnection == nil {
		return nil
	}
	closeCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return databaseConnection.Close(closeCtx)
}
