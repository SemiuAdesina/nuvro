package clock_test

import (
	"testing"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
)

func TestSystemReturnsUTC(t *testing.T) {
	now := clock.NewSystem().Now()

	if now.Location() != time.UTC {
		t.Fatalf("Location() = %v, want UTC", now.Location())
	}
	if time.Since(now) > time.Minute {
		t.Fatalf("Now() = %v, which is not close to the current instant", now)
	}
}

func TestFixedNormalisesToUTCAndDoesNotDrift(t *testing.T) {
	instant := time.Date(2026, time.September, 7, 9, 0, 0, 0, time.FixedZone("WAT", 3600))
	fixed := clock.NewFixed(instant)

	first := fixed.Now()
	second := fixed.Now()

	if !first.Equal(second) {
		t.Fatalf("Now() drifted between calls: %v and %v", first, second)
	}
	if first.Location() != time.UTC {
		t.Fatalf("Location() = %v, want UTC", first.Location())
	}
	if !first.Equal(instant) {
		t.Fatalf("Now() = %v, want the same instant as %v", first, instant)
	}
}

func TestFixedAdvanceMovesTimeForward(t *testing.T) {
	fixed := clock.NewFixed(time.Date(2026, time.September, 7, 9, 0, 0, 0, time.UTC))

	fixed.Advance(90 * time.Minute)

	want := time.Date(2026, time.September, 7, 10, 30, 0, 0, time.UTC)
	if !fixed.Now().Equal(want) {
		t.Fatalf("Now() = %v, want %v", fixed.Now(), want)
	}
}

func TestFixedSatisfiesClock(t *testing.T) {
	var _ clock.Clock = clock.NewFixed(time.Now())
	var _ clock.Clock = clock.NewSystem()
}
