package main

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"majucau.local/financial-intelligence/internal/integrations/googledrive"
	"majucau.local/financial-intelligence/internal/scheduler"
)

func startDailyGoogleDriveSchedule(ctx context.Context, syncer dailyGoogleDriveSyncer, configurationErrorCode string, logger *slog.Logger) <-chan struct{} {
	done := make(chan struct{})
	if syncer == nil {
		if configurationErrorCode == "" {
			configurationErrorCode = "GOOGLE_DRIVE_NOT_CONFIGURED"
		}
		logger.Info("daily Google Drive inventory is pending configuration", slog.String("error_code", configurationErrorCode))
		close(done)
		return done
	}
	go func() {
		defer close(done)
		schedule, err := scheduler.NewSaoPaulo0930()
		if err != nil {
			logger.Error("daily Google Drive schedule is unavailable", slog.String("error_code", "DAILY_SCHEDULE_INVALID"))
			return
		}
		logger.Info("daily Google Drive inventory schedule started", slog.String("time_zone", scheduler.SaoPauloTimeZone), slog.Int("hour", scheduler.DailySyncHour), slog.Int("minute", scheduler.DailySyncMinute))
		if scheduledAt, due := dailyGoogleDriveCatchUpAt(time.Now()); due {
			result, syncErr := runDailyGoogleDriveSync(ctx, syncer, scheduledAt)
			if syncErr != nil {
				logger.Warn("daily Google Drive recovery did not complete", dailyGoogleDriveResultLogFields(scheduledAt, result, syncErr)...)
			} else {
				logDailyGoogleDriveSuccess(logger, scheduledAt, result)
			}
		}
		err = schedule.Run(ctx, func(runCtx context.Context, scheduledAt time.Time) error {
			result, syncErr := runDailyGoogleDriveSync(runCtx, syncer, scheduledAt)
			if syncErr != nil {
				logger.Warn("daily Google Drive processing did not complete", dailyGoogleDriveResultLogFields(scheduledAt, result, syncErr)...)
				return syncErr
			}
			logDailyGoogleDriveSuccess(logger, scheduledAt, result)
			return nil
		}, func(scheduledAt time.Time, runErr error) {
			logger.Warn("daily Google Drive inventory did not complete", slog.String("scheduled_at", scheduledAt.Format(time.RFC3339)), slog.String("error_code", dailyGoogleDriveSyncErrorCode(runErr)))
		})
		if err != nil && !errors.Is(err, context.Canceled) {
			logger.Error("daily Google Drive schedule stopped", slog.String("error_code", "DAILY_SCHEDULE_FAILED"))
		}
	}()
	return done
}

func runDailyGoogleDriveSync(ctx context.Context, syncer dailyGoogleDriveSyncer, scheduledAt time.Time) (googledrive.DailyFileInventorySyncResult, error) {
	if syncer == nil {
		return googledrive.DailyFileInventorySyncResult{Status: "FAILED"}, googledrive.ErrSyncConfiguration
	}
	result, err := syncer.SyncDaily(ctx, scheduledAt)
	if err != nil {
		return result, err
	}
	if result.Status != "SUCCESS" {
		return result, googledrive.ErrDailySyncIncomplete
	}
	return result, nil
}

func dailyGoogleDriveCatchUpAt(now time.Time) (time.Time, bool) {
	location, err := time.LoadLocation(scheduler.SaoPauloTimeZone)
	if err != nil {
		return time.Time{}, false
	}
	local := now.In(location)
	year, month, day := local.Date()
	scheduledAt := time.Date(year, month, day, scheduler.DailySyncHour, scheduler.DailySyncMinute, 0, 0, location)
	return scheduledAt, !local.Before(scheduledAt)
}

func logDailyGoogleDriveSuccess(logger *slog.Logger, scheduledAt time.Time, result googledrive.DailyFileInventorySyncResult) {
	logger.Info("daily Google Drive processing completed", dailyGoogleDriveResultLogFields(scheduledAt, result, nil)...)
}

func dailyGoogleDriveResultLogFields(scheduledAt time.Time, result googledrive.DailyFileInventorySyncResult, syncErr error) []any {
	fields := []any{
		slog.String("scheduled_at", scheduledAt.Format(time.RFC3339)),
		slog.String("status", result.Status),
		slog.Int("folders_read", result.FoldersRead),
		slog.Int("folders_failed", result.FoldersFailed),
		slog.Int("folders_skipped", result.FoldersSkipped),
		slog.Int64("files_read", result.FilesRead),
		slog.String("future_import_status", result.FutureStatus),
		slog.Int64("future_records_read", result.FutureRead),
		slog.Int64("future_records_created", result.FutureCreated),
		slog.Int64("future_records_updated", result.FutureUpdated),
		slog.Int64("future_records_failed", result.FutureFailed),
		slog.Int64("future_files_pending", result.FuturePending),
		slog.Int64("future_files_stale", result.FutureStale),
	}
	if syncErr != nil {
		fields = append(fields, slog.String("error_code", dailyGoogleDriveSyncErrorCode(syncErr)))
	}
	return fields
}

func dailyGoogleDriveSyncErrorCode(err error) string {
	switch {
	case errors.Is(err, googledrive.ErrDailySyncIncomplete):
		return "GOOGLE_DRIVE_SYNC_INCOMPLETE"
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return "GOOGLE_DRIVE_SYNC_CANCELLED"
	default:
		return "GOOGLE_DRIVE_SYNC_FAILED"
	}
}
