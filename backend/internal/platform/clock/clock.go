package clock

import "time"

type Clock interface {
	Now() time.Time
}

type System struct{}

func NewSystem() System {
	return System{}
}

func (System) Now() time.Time {
	return time.Now().UTC()
}

type Fixed struct {
	instant time.Time
}

func NewFixed(instant time.Time) *Fixed {
	return &Fixed{instant: instant.UTC()}
}

func (f *Fixed) Now() time.Time {
	return f.instant
}

func (f *Fixed) Advance(by time.Duration) {
	f.instant = f.instant.Add(by)
}
