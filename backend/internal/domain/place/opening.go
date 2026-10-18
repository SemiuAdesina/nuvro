package place

import "time"

type OpeningStatus string

const (
	Open          OpeningStatus = "open"
	Closed        OpeningStatus = "closed"
	UnknownStatus OpeningStatus = "unknown"
)

type Interval struct {
	Weekday time.Weekday
	Opens   int
	Closes  int
}

type OpeningHours struct {
	Intervals  []Interval
	ObservedAt time.Time
	Known      bool
}

func (i Interval) Valid() bool {
	return i.Opens >= 0 && i.Closes > i.Opens && i.Closes <= 24*60 && i.Weekday >= time.Sunday && i.Weekday <= time.Saturday
}

func (h OpeningHours) OpenDuring(start, end time.Time) OpeningStatus {
	if !h.Known || len(h.Intervals) == 0 || end.Before(start) {
		return UnknownStatus
	}
	if start.YearDay() != end.YearDay() || start.Year() != end.Year() {
		return UnknownStatus
	}
	from := start.Hour()*60 + start.Minute()
	until := end.Hour()*60 + end.Minute()
	for _, interval := range h.Intervals {
		if interval.Weekday != start.Weekday() || !interval.Valid() {
			continue
		}
		if from >= interval.Opens && until <= interval.Closes {
			return Open
		}
	}
	return Closed
}
