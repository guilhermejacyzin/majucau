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

func startDailyBlingSchedule(ctx context.Context, syncer blingDailySyncer, logger *slog.Logger) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		defer close(done)
		schedule, err := scheduler.NewSaoPaulo0930()
		if err != nil {
			logger.Error("daily synchronization schedule is unavailable", slog.String("error_code", "DAILY_SCHEDULE_INVALID"))
			return
		}
		logger.Info("daily Bling synchronization schedule started", slog.String("time_zone", scheduler.SaoPauloTimeZone), slog.Int("hour", scheduler.DailySyncHour), slog.Int("minute", scheduler.DailySyncMinute))
		if scheduledAt, due := dailyBlingCatchUpAt(time.Now()); due {
			result, skipped, runErr := runDailyBlingSync(ctx, syncer, scheduledAt)
			if runErr != nil {
				logger.Warn("daily Bling recovery did not complete", slog.String("scheduled_at", scheduledAt.Format(time.RFC3339)), slog.String("error_code", dailyBlingSyncErrorCode(runErr)))
			} else if skipped {
				logger.Info("daily Bling synchronization already completed", slog.String("scheduled_at", scheduledAt.Format(time.RFC3339)))
			} else {
				logDailyBlingSuccess(logger, scheduledAt, result)
			}
		}
		err = schedule.Run(ctx, func(runCtx context.Context, scheduledAt time.Time) error {
			result, skipped, runErr := runDailyBlingSync(runCtx, syncer, scheduledAt)
			if runErr != nil {
				return runErr
			}
			if skipped {
				logger.Info("daily Bling synchronization already completed", slog.String("scheduled_at", scheduledAt.Format(time.RFC3339)))
				return nil
			}
			logDailyBlingSuccess(logger, scheduledAt, result)
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

func runDailyBlingSync(ctx context.Context, syncer blingDailySyncer, scheduledAt time.Time) (bling.BlingRawSyncResult, bool, error) {
	if syncer == nil {
		return bling.BlingRawSyncResult{}, false, ErrDailyBlingUnavailable
	}
	result, skipped, err := syncer.SyncDaily(ctx, scheduledAt)
	if err != nil {
		return result, false, err
	}
	if result.Status != "SUCCESS" {
		return result, false, ErrDailyBlingIncomplete
	}
	return result, skipped, nil
}

func dailyBlingCatchUpAt(now time.Time) (time.Time, bool) {
	location, err := time.LoadLocation(scheduler.SaoPauloTimeZone)
	if err != nil {
		return time.Time{}, false
	}
	local := now.In(location)
	year, month, day := local.Date()
	scheduledAt := time.Date(year, month, day, scheduler.DailySyncHour, scheduler.DailySyncMinute, 0, 0, location)
	return scheduledAt, !local.Before(scheduledAt)
}

func logDailyBlingSuccess(logger *slog.Logger, scheduledAt time.Time, result bling.BlingRawSyncResult) {
	logger.Info("daily Bling synchronization completed", slog.String("scheduled_at", scheduledAt.Format(time.RFC3339)), slog.Int("pages_read", result.Receivables.PagesRead+result.Payables.PagesRead), slog.Int("records_read", result.Receivables.RecordsRead+result.Payables.RecordsRead))
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
