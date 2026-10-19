package scheduler

import (
	"context"
	"errors"
	"time"
	_ "time/tzdata"

	"github.com/google/uuid"
)

const (
	SaoPauloTimeZone   = "America/Sao_Paulo"
	DailySyncHour      = 9
	DailySyncMinute    = 30
	BlingDailyRunScope = "majucau/bling/daily-sync/v1"
)

var (
	ErrInvalidDailySchedule = errors.New("invalid daily schedule")
	ErrDailyActionRequired  = errors.New("daily schedule action is required")
)

// Daily runs an action once at a fixed local time each day. NextRun is always
// strictly after the supplied instant, so starting the worker after the daily
// time does not trigger a same-day catch-up that could read late-arriving files.
type Daily struct {
	location *time.Location
	hour     int
	minute   int
	now      func() time.Time
	wait     func(context.Context, time.Duration) error
}

func NewDaily(timeZone string, hour, minute int) (Daily, error) {
	if hour < 0 || hour > 23 || minute < 0 || minute > 59 || timeZone == "" {
		return Daily{}, ErrInvalidDailySchedule
	}
	location, err := time.LoadLocation(timeZone)
	if err != nil {
		return Daily{}, errors.Join(ErrInvalidDailySchedule, err)
	}
	return Daily{location: location, hour: hour, minute: minute, now: time.Now, wait: waitContext}, nil
}

func NewSaoPaulo0930() (Daily, error) {
	return NewDaily(SaoPauloTimeZone, DailySyncHour, DailySyncMinute)
}

func (d Daily) NextRun(after time.Time) (time.Time, error) {
	if d.location == nil || d.hour < 0 || d.hour > 23 || d.minute < 0 || d.minute > 59 {
		return time.Time{}, ErrInvalidDailySchedule
	}
	local := after.In(d.location)
	year, month, day := local.Date()
	next := time.Date(year, month, day, d.hour, d.minute, 0, 0, d.location)
	if !next.After(local) {
		following := local.AddDate(0, 0, 1)
		year, month, day = following.Date()
		next = time.Date(year, month, day, d.hour, d.minute, 0, 0, d.location)
	}
	return next, nil
}

// LatestRunAtOrBefore returns the most recent scheduled instant at or before
// at. It lets readers distinguish today's completed load from yesterday's
// snapshot after the daily update time has passed.
func (d Daily) LatestRunAtOrBefore(at time.Time) (time.Time, error) {
	if d.location == nil || d.hour < 0 || d.hour > 23 || d.minute < 0 || d.minute > 59 {
		return time.Time{}, ErrInvalidDailySchedule
	}
	local := at.In(d.location)
	year, month, day := local.Date()
	latest := time.Date(year, month, day, d.hour, d.minute, 0, 0, d.location)
	if latest.After(local) {
		previous := local.AddDate(0, 0, -1)
		year, month, day = previous.Date()
		latest = time.Date(year, month, day, d.hour, d.minute, 0, 0, d.location)
	}
	return latest, nil
}

// DailyRunCorrelationID creates a stable identifier for one source's
// scheduled run. The identity format is versioned because persisted batches
// use this value to prove that the full scheduled load completed.
func DailyRunCorrelationID(scope string, scheduledAt time.Time) uuid.UUID {
	return uuid.NewSHA1(uuid.NameSpaceOID, []byte(scope+":"+scheduledAt.Format("2006-01-02")))
}

// Run waits for the next scheduled instant, executes action, reports an action
// error, and continues to the next day. A failed daily run is not retried
// outside the API's bounded request retry policy or the next scheduled run.
func (d Daily) Run(
	ctx context.Context,
	action func(context.Context, time.Time) error,
	onError func(time.Time, error),
) error {
	if ctx == nil || action == nil || d.location == nil || d.now == nil || d.wait == nil {
		return ErrInvalidDailySchedule
	}
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		now := d.now()
		next, err := d.NextRun(now)
		if err != nil {
			return err
		}
		if err := d.wait(ctx, next.Sub(now)); err != nil {
			return err
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := action(ctx, next); err != nil && ctx.Err() == nil && onError != nil {
			onError(next, err)
		}
	}
}

func waitContext(ctx context.Context, duration time.Duration) error {
	if duration <= 0 {
		return ErrInvalidDailySchedule
	}
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
