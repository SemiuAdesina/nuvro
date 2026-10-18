package evidence

import "time"

type Status string

const (
	Live   Status = "live"
	Cached Status = "cached"
	Stale  Status = "stale"
)

type Freshness struct {
	ObservedAt time.Time
	ValidUntil time.Time
	Status     Status
}

type Source struct {
	Provider  string
	Reference string
	Freshness Freshness
}
