package domain

import (
	"time"
	_ "time/tzdata"
)

var saoPauloLocation = func() *time.Location {
	location, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		panic("load America/Sao_Paulo business timezone: " + err.Error())
	}
	return location
}()

// BusinessDate converts a technical instant to the Sao Paulo business date.
// The result uses midnight UTC only as a stable date-only representation.
func BusinessDate(instant time.Time) time.Time {
	if instant.IsZero() {
		return time.Time{}
	}
	year, month, day := instant.In(saoPauloLocation).Date()
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}
