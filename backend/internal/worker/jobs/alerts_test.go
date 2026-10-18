package jobs_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/alert"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/itinerary"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/place"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/trip"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/weather"
	"github.com/SemiuAdesina/nuvro/backend/internal/engine/alerting"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
	"github.com/SemiuAdesina/nuvro/backend/internal/worker/jobs"
)

var now = time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC)

type source struct {
	trips []trip.Trip
	err   error
	from  time.Time
	until time.Time
	limit int
}

func (s *source) TripsNeedingEvaluation(_ context.Context, from, until time.Time, limit int) ([]trip.Trip, error) {
	s.from, s.until, s.limit = from, until, limit
	return s.trips, s.err
}

type snapshots struct {
	inputs map[string]alerting.Input
	seen   []string
}

func (s *snapshots) Evaluate(_ context.Context, tripID string) (alerting.Input, bool) {
	s.seen = append(s.seen, tripID)
	input, found := s.inputs[tripID]
	return input, found
}

type recorder struct {
	recorded []alert.Alert
	err      error
}

func (r *recorder) Record(_ context.Context, produced []alert.Alert) ([]alert.Alert, error) {
	if r.err != nil {
		return nil, r.err
	}
	r.recorded = append(r.recorded, produced...)
	return produced, nil
}

const fixtureTripID = "3f1c9c8e-0e3a-4f0a-9b1f-1a2b3c4d5e6f"

func rainyTrip() alerting.Input {
	garden := itinerary.Item{
		ID: "garden", TripID: fixtureTripID, Day: 3, StartAt: now.Add(3 * time.Hour), EndAt: now.Add(5 * time.Hour),
		Kind: itinerary.Activity, Title: "Garden", PlaceID: "place-garden", Category: place.Park,
		Status: itinerary.Planned, Source: itinerary.Generated,
	}
	return alerting.Input{
		TripID: fixtureTripID, Timezone: "Atlantic/Madeira", Now: now,
		Itinerary: itinerary.Itinerary{TripID: fixtureTripID, Items: []itinerary.Item{garden}},
		Weather: weather.Window{
			Availability: weather.Available,
			Hourly: []weather.Observation{{
				At: garden.StartAt, Condition: weather.HeavyRain, TemperatureCelsius: 18, PrecipitationChance: 90,
			}},
		},
	}
}

func newJob(t *testing.T, trips *source, states *snapshots, store *recorder) *jobs.AlertEvaluation {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return jobs.NewAlertEvaluation(trips, states, alerting.New(), store, logger, clock.NewFixed(now), 0)
}

func TestJobEvaluatesActiveTripsAndRecordsAlerts(t *testing.T) {
	tripID := "3f1c9c8e-0e3a-4f0a-9b1f-1a2b3c4d5e6f"
	trips := &source{trips: []trip.Trip{{ID: tripID}}}
	states := &snapshots{inputs: map[string]alerting.Input{tripID: rainyTrip()}}
	store := &recorder{}
	if err := newJob(t, trips, states, store).Once(t.Context()); err != nil {
		t.Fatal(err)
	}
	if len(store.recorded) == 0 {
		t.Fatal("no alerts were recorded")
	}
	if store.recorded[0].Type != alert.WeatherImpact {
		t.Fatalf("alert = %+v", store.recorded[0])
	}
}

func TestJobOnlyLooksAtTripsInsideItsHorizon(t *testing.T) {
	trips := &source{}
	if err := newJob(t, trips, &snapshots{}, &recorder{}).Once(t.Context()); err != nil {
		t.Fatal(err)
	}
	if !trips.from.Equal(now) {
		t.Fatalf("window started at %v", trips.from)
	}
	if trips.until.Sub(trips.from) > 24*time.Hour || trips.until.Before(trips.from) {
		t.Fatalf("window is %v wide", trips.until.Sub(trips.from))
	}
	if trips.limit <= 0 || trips.limit > 1000 {
		t.Fatalf("limit = %d", trips.limit)
	}
}

func TestTripsWithoutUsableStateAreSkipped(t *testing.T) {
	tripID := "3f1c9c8e-0e3a-4f0a-9b1f-1a2b3c4d5e6f"
	trips := &source{trips: []trip.Trip{{ID: tripID}}}
	states := &snapshots{inputs: map[string]alerting.Input{}}
	store := &recorder{}
	if err := newJob(t, trips, states, store).Once(t.Context()); err != nil {
		t.Fatal(err)
	}
	if len(store.recorded) != 0 {
		t.Fatal("alerts were produced for a trip with no state")
	}
	if len(states.seen) != 1 {
		t.Fatal("the trip was not evaluated")
	}
}

func TestNothingWorthAlertingProducesNoWrite(t *testing.T) {
	tripID := "3f1c9c8e-0e3a-4f0a-9b1f-1a2b3c4d5e6f"
	quiet := rainyTrip()
	quiet.Weather.Hourly[0].Condition = weather.Clear
	trips := &source{trips: []trip.Trip{{ID: tripID}}}
	store := &recorder{}
	if err := newJob(t, trips, &snapshots{inputs: map[string]alerting.Input{tripID: quiet}}, store).
		Once(t.Context()); err != nil {
		t.Fatal(err)
	}
	if len(store.recorded) != 0 {
		t.Fatalf("recorded %v", store.recorded)
	}
}

func TestFailuresSurfaceRatherThanSilentlyStopping(t *testing.T) {
	tripID := "3f1c9c8e-0e3a-4f0a-9b1f-1a2b3c4d5e6f"
	if err := newJob(t, &source{err: errors.New("database down")}, &snapshots{}, &recorder{}).
		Once(t.Context()); err == nil {
		t.Fatal("a failing trip query reported success")
	}
	trips := &source{trips: []trip.Trip{{ID: tripID}}}
	states := &snapshots{inputs: map[string]alerting.Input{tripID: rainyTrip()}}
	if err := newJob(t, trips, states, &recorder{err: errors.New("write failed")}).
		Once(t.Context()); err == nil {
		t.Fatal("a failing recorder reported success")
	}
}

func TestJobStopsOnCancellation(t *testing.T) {
	tripID := "3f1c9c8e-0e3a-4f0a-9b1f-1a2b3c4d5e6f"
	trips := &source{trips: []trip.Trip{{ID: tripID}}}
	states := &snapshots{inputs: map[string]alerting.Input{tripID: rainyTrip()}}
	job := newJob(t, trips, states, &recorder{})
	if job.Name() != "alert-evaluation" {
		t.Fatalf("name = %q", job.Name())
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if err := job.Run(ctx); err != nil {
		t.Fatal(err)
	}
}
