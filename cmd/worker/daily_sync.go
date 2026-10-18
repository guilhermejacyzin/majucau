package main

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"majucau.local/financial-intelligence/internal/integrations/bling"
	"majucau.local/financial-intelligence/internal/scheduler"
)

var (
	ErrDailyBlingUnavailable = errors.New("daily Bling sync is not configured")
	ErrDailyBlingIncomplete  = errors.New("daily Bling sync did not complete")
)

func startDailyBlingSchedule(ctx context.Context, syncer blingRawSyncer, logger *slog.Logger) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		defer close(done)
		schedule, err := scheduler.NewSaoPaulo0930()
		if err != nil {
			logger.Error("daily synchronization schedule is unavailable", slog.String("error_code", "DAILY_SCHEDULE_INVALID"))
			return
		}
		logger.Info("daily Bling synchronization schedule started", slog.String("time_zone", scheduler.SaoPauloTimeZone), slog.Int("hour", scheduler.DailySyncHour), slog.Int("minute", scheduler.DailySyncMinute))
		err = schedule.Run(ctx, func(runCtx context.Context, scheduledAt time.Time) error {
			result, runErr := runDailyBlingSync(runCtx, syncer)
			if runErr != nil {
				return runErr
			}
			logger.Info("daily Bling synchronization completed", slog.String("scheduled_at", scheduledAt.Format(time.RFC3339)), slog.Int("pages_read", result.Receivables.PagesRead+result.Payables.PagesRead), slog.Int("records_read", result.Receivables.RecordsRead+result.Payables.RecordsRead))
			return nil
		}, func(_ time.Time, runErr error) {
			logger.Warn("daily Bling synchronization did not complete", slog.String("error_code", dailyBlingSyncErrorCode(runErr)))
		})
		if err != nil && !errors.Is(err, context.Canceled) {
			logger.Error("daily synchronization schedule stopped", slog.String("error_code", "DAILY_SCHEDULE_FAILED"))
		}
	}()
	return done
}

func runDailyBlingSync(ctx context.Context, syncer blingRawSyncer) (bling.BlingRawSyncResult, error) {
	if syncer == nil {
		return bling.BlingRawSyncResult{}, ErrDailyBlingUnavailable
	}
	result, err := syncer.Sync(ctx, bling.ReceivablesFilter{})
	if err != nil {
		return result, err
	}
	if result.Status != "SUCCESS" {
		return result, ErrDailyBlingIncomplete
	}
	return result, nil
}

func dailyBlingSyncErrorCode(err error) string {
	switch {
	case errors.Is(err, ErrDailyBlingUnavailable):
		return "BLING_CONNECTION_NOT_CONFIGURED"
	case errors.Is(err, ErrDailyBlingIncomplete):
		return "BLING_SYNC_INCOMPLETE"
	default:
		return blingSyncErrorCode(err)
	}
}
