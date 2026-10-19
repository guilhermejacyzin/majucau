package scheduler

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNextRunUsesSaoPauloTimeAndWaitsForTomorrowAtBoundary(t *testing.T) {
	schedule, err := NewSaoPaulo0930()
	if err != nil {
		t.Fatal(err)
	}

	before := time.Date(2026, time.October, 18, 12, 29, 59, 0, time.UTC)
	got, err := schedule.NextRun(before)
	if err != nil {
		t.Fatal(err)
	}
	if want := "2026-10-18T09:30:00-03:00"; got.Format(time.RFC3339) != want {
		t.Fatalf("next run before boundary = %s, want %s", got.Format(time.RFC3339), want)
	}

	atBoundary := time.Date(2026, time.October, 18, 12, 30, 0, 0, time.UTC)
	got, err = schedule.NextRun(atBoundary)
	if err != nil {
		t.Fatal(err)
	}
	if want := "2026-10-19T09:30:00-03:00"; got.Format(time.RFC3339) != want {
		t.Fatalf("next run at boundary = %s, want %s", got.Format(time.RFC3339), want)
	}
}

func TestRunReportsFailureAndContinuesOnFollowingDay(t *testing.T) {
	schedule, err := NewSaoPaulo0930()
	if err != nil {
		t.Fatal(err)
	}
	current := time.Date(2026, time.October, 18, 12, 29, 0, 0, time.UTC)
	schedule.now = func() time.Time { return current }
	schedule.wait = func(ctx context.Context, duration time.Duration) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		current = current.Add(duration)
		return nil
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var ran []time.Time
	var reported []error
	gotErr := schedule.Run(ctx, func(_ context.Context, scheduled time.Time) error {
		ran = append(ran, scheduled)
		if len(ran) == 1 {
			return errors.New("temporary failure")
		}
		cancel()
		return nil
	}, func(_ time.Time, err error) {
		reported = append(reported, err)
	})
	if !errors.Is(gotErr, context.Canceled) {
		t.Fatalf("Run error = %v, want context cancellation", gotErr)
	}
	if len(ran) != 2 || len(reported) != 1 {
		t.Fatalf("runs=%d reported failures=%d, want 2 runs and 1 failure", len(ran), len(reported))
	}
	if got := ran[0].In(ran[0].Location()).Format("2006-01-02 15:04"); got != "2026-10-18 09:30" {
		t.Fatalf("first scheduled run = %s", got)
	}
	if got := ran[1].In(ran[1].Location()).Format("2006-01-02 15:04"); got != "2026-10-19 09:30" {
		t.Fatalf("second scheduled run = %s", got)
	}
}

func TestNewDailyRejectsInvalidTimeZoneAndClock(t *testing.T) {
	if _, err := NewDaily("not/a/timezone", 9, 30); !errors.Is(err, ErrInvalidDailySchedule) {
		t.Fatalf("invalid timezone error = %v", err)
	}
	if _, err := NewDaily(SaoPauloTimeZone, 24, 0); !errors.Is(err, ErrInvalidDailySchedule) {
		t.Fatalf("invalid hour error = %v", err)
	}
}

func TestLatestRunAtOrBeforeUsesTheMostRecentBrasiliaSlot(t *testing.T) {
	schedule, err := NewSaoPaulo0930()
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name string
		at   time.Time
		want string
	}{
		{name: "before today's slot", at: time.Date(2026, 10, 19, 12, 29, 0, 0, time.UTC), want: "2026-10-18T09:30:00-03:00"},
		{name: "at today's slot", at: time.Date(2026, 10, 19, 12, 30, 0, 0, time.UTC), want: "2026-10-19T09:30:00-03:00"},
		{name: "after today's slot", at: time.Date(2026, 10, 19, 15, 0, 0, 0, time.UTC), want: "2026-10-19T09:30:00-03:00"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := schedule.LatestRunAtOrBefore(tc.at)
			if err != nil {
				t.Fatal(err)
			}
			if formatted := got.Format(time.RFC3339); formatted != tc.want {
				t.Fatalf("latest slot = %s, want %s", formatted, tc.want)
			}
		})
	}
}

func TestDailyRunCorrelationIDPreservesExistingBlingIdentity(t *testing.T) {
	slot := time.Date(2026, 10, 19, 9, 30, 0, 0, time.FixedZone("Brasilia", -3*60*60))
	want := uuid.NewSHA1(uuid.NameSpaceOID, []byte("majucau/bling/daily-sync/v1:2026-10-19"))
	if got := DailyRunCorrelationID(BlingDailyRunScope, slot); got != want {
		t.Fatalf("daily run ID = %s, want existing ID %s", got, want)
	}
}
