package domain

import (
	"testing"
	"time"
)

func TestBusinessDateUsesSaoPauloCalendar(t *testing.T) {
	cases := []struct {
		instant time.Time
		want    time.Time
	}{
		{
			instant: time.Date(2026, 3, 1, 2, 59, 0, 0, time.UTC),
			want:    time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC),
		},
		{
			instant: time.Date(2026, 3, 1, 3, 0, 0, 0, time.UTC),
			want:    time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
		},
	}
	for _, tc := range cases {
		if got := BusinessDate(tc.instant); !got.Equal(tc.want) || got.Location() != time.UTC {
			t.Errorf("BusinessDate(%s) = %s (%s); want date %s at UTC midnight", tc.instant, got, got.Location(), tc.want.Format("2006-01-02"))
		}
	}
}
