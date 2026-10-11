package financial

import (
	"context"
	"errors"
	"time"

	"majucau.local/financial-intelligence/internal/domain"
)

type AgingBucket string

const (
	Overdue        AgingBucket = "OVERDUE"
	Today          AgingBucket = "TODAY"
	Days1To7       AgingBucket = "D1_7"
	Days8To15      AgingBucket = "D8_15"
	Days16To30     AgingBucket = "D16_30"
	Days31To45     AgingBucket = "D31_45"
	Days46To60     AgingBucket = "D46_60"
	OutsideHorizon AgingBucket = "OUTSIDE_HORIZON"
)

type OpenReceivable struct {
	DueDate     time.Time
	OpenBalance domain.Money
	Status      string
}
type AgingTotal struct {
	Bucket AgingBucket
	Amount domain.Money
	Count  int
}

// OpenReceivableReader yields one open receivable at a time. Database-backed
// callers should use this contract instead of materializing the full result.
type OpenReceivableReader func(context.Context) (OpenReceivable, bool, error)

func AgingBucketFor(reference, due time.Time) AgingBucket {
	// D-005-A treats the reference as an instant and due as a persisted business
	// date. Convert only the instant; converting a SQL DATE would shift its day.
	referenceDate := domain.BusinessDate(reference)
	dueDate := dateOnly(due)
	days := int(dueDate.Sub(referenceDate).Hours() / 24)
	switch {
	case days < 0:
		return Overdue
	case days == 0:
		return Today
	case days <= 7:
		return Days1To7
	case days <= 15:
		return Days8To15
	case days <= 30:
		return Days16To30
	case days <= 45:
		return Days31To45
	case days <= 60:
		return Days46To60
	default:
		return OutsideHorizon
	}
}

func CalculateAging(reference time.Time, records []OpenReceivable) (map[AgingBucket]AgingTotal, error) {
	index := 0
	return CalculateAgingFrom(context.Background(), reference, func(context.Context) (OpenReceivable, bool, error) {
		if index >= len(records) {
			return OpenReceivable{}, false, nil
		}
		record := records[index]
		index++
		return record, true, nil
	})
}

// CalculateAgingFrom consumes receivables incrementally and retains only the
// fixed set of aging totals. Reader, validation, and cancellation errors return
// no partial totals. It preserves the same D-005-A buckets and exclusions as
// CalculateAging.
func CalculateAgingFrom(ctx context.Context, reference time.Time, next OpenReceivableReader) (map[AgingBucket]AgingTotal, error) {
	if reference.IsZero() {
		return nil, errors.New("reference date is required")
	}
	if ctx == nil {
		return nil, errors.New("context is required")
	}
	if next == nil {
		return nil, errors.New("receivable reader is required")
	}
	result := make(map[AgingBucket]AgingTotal)
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		r, ok, err := next(ctx)
		if err != nil {
			return nil, err
		}
		if !ok {
			break
		}
		status := r.Status
		if status == "CANCELLED" || status == "CANCELED" || status == "REFUNDED" || status == "REVERSED" || status == "ESTORNADO" {
			continue
		}
		if r.OpenBalance.IsNegative() {
			return nil, errors.New("open balance cannot be negative")
		}
		if r.DueDate.IsZero() {
			return nil, errors.New("due date is required")
		}
		if r.OpenBalance.IsZero() {
			continue
		}
		bucket := AgingBucketFor(reference, r.DueDate)
		item := result[bucket]
		item.Amount, err = item.Amount.Add(r.OpenBalance)
		if err != nil {
			return nil, err
		}
		item.Count++
		item.Bucket = bucket
		result[bucket] = item
	}
	return result, nil
}
