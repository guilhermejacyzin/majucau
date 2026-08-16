package financial

import (
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

func AgingBucketFor(reference, due time.Time) AgingBucket {
	days := int(dateOnly(due).Sub(dateOnly(reference)).Hours() / 24)
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
	if reference.IsZero() {
		return nil, errors.New("reference date is required")
	}
	result := make(map[AgingBucket]AgingTotal)
	for _, r := range records {
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
		var err error
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
