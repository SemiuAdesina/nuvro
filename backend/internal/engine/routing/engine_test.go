package routing_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/geo"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/route"
	"github.com/SemiuAdesina/nuvro/backend/internal/engine/routing"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
)

var now = time.Date(2026, 10, 14, 9, 0, 0, 0, time.UTC)
var origin = geo.Point{Latitude: 32.6478, Longitude: -16.9048}
var destination = geo.Point{Latitude: 32.6690, Longitude: -16.9000}

type provider struct {
	evidence route.Evidence
	matrix   route.Matrix
	err      error
	calls    int
}

func (p *provider) Compute(context.Context, routing.Query) (route.Evidence, error) {
	p.calls++
	return p.evidence, p.err
}

func (p *provider) Matrix(_ context.Context, query routing.MatrixQuery) (route.Matrix, error) {
	p.calls++
	matrix := p.matrix
	matrix.Points = query.Points
	return matrix, p.err
}

type store struct {
	entries map[string]route.Evidence
	ttl     time.Duration
}

func (s *store) Route(_ context.Context, key string) (route.Evidence, bool) {
	value, found := s.entries[key]
	return value, found
}

func (s *store) StoreRoute(_ context.Context, key string, evidence route.Evidence, ttl time.Duration) {
	if s.entries == nil {
		s.entries = map[string]route.Evidence{}
	}
	s.entries[key] = evidence
	s.ttl = ttl
}

func found(mode route.Mode, seconds int) route.Evidence {
	return route.Evidence{
		Status: route.Found, Origin: origin, Destination: destination, Mode: mode,
		DurationSeconds: seconds, DistanceMeters: seconds * 2, Confidence: route.High, ComputedAt: now,
	}
}

func TestComputeReturnsProviderEvidence(t *testing.T) {
	source := &provider{evidence: found(route.Walk, 1200)}
	evidence := routing.New(source, nil, clock.NewFixed(now)).
		Compute(t.Context(), routing.Query{Origin: origin, Destination: destination, Mode: route.Walk})
	if !evidence.Known() || evidence.Duration() != 20*time.Minute {
		t.Fatalf("evidence = %+v", evidence)
	}
}

func TestUnconfiguredOrFailingProviderNeverInventsATime(t *testing.T) {
	unconfigured := routing.New(nil, nil, clock.NewFixed(now))
	if unconfigured.Configured() {
		t.Fatal("an unconfigured engine claimed to be configured")
	}
	evidence := unconfigured.Compute(t.Context(), routing.Query{Origin: origin, Destination: destination, Mode: route.Walk})
	if evidence.Known() || evidence.Status != route.Unavailable {
		t.Fatalf("unconfigured evidence = %+v", evidence)
	}
	failing := routing.New(&provider{err: errors.New("provider down")}, nil, clock.NewFixed(now))
	if failing.Compute(t.Context(), routing.Query{Origin: origin, Destination: destination, Mode: route.Drive}).Known() {
		t.Fatal("a failing provider produced a travel time")
	}
}

func TestInvalidModesAreRefused(t *testing.T) {
	source := &provider{evidence: found(route.Walk, 600)}
	evidence := routing.New(source, nil, clock.NewFixed(now)).
		Compute(t.Context(), routing.Query{Origin: origin, Destination: destination, Mode: "teleport"})
	if evidence.Known() || source.calls != 0 {
		t.Fatal("an unsupported mode reached the provider")
	}
}

func TestWalkingRoutesCacheLongerThanTrafficRoutes(t *testing.T) {
	if routing.TimeToLive(route.Walk) <= routing.TimeToLive(route.Drive) {
		t.Fatal("walking should outlive traffic-dependent driving")
	}
	if routing.TimeToLive(route.Bike) != routing.StaticCacheTTL || routing.TimeToLive(route.Transit) != routing.TrafficCacheTTL {
		t.Fatal("cache lifetimes are not classified by volatility")
	}
	cache := &store{}
	source := &provider{evidence: found(route.Walk, 900)}
	engine := routing.New(source, cache, clock.NewFixed(now))
	query := routing.Query{Origin: origin, Destination: destination, Mode: route.Walk}
	engine.Compute(t.Context(), query)
	engine.Compute(t.Context(), query)
	if source.calls != 1 {
		t.Fatalf("provider called %d times for an identical walking route", source.calls)
	}
	if cache.ttl != routing.StaticCacheTTL {
		t.Fatalf("walking ttl = %v", cache.ttl)
	}
}

func TestDrivingRoutesAreNotReusedAcrossUnrelatedTimes(t *testing.T) {
	cache := &store{}
	source := &provider{evidence: found(route.Drive, 600)}
	engine := routing.New(source, cache, clock.NewFixed(now))
	morning := routing.Query{Origin: origin, Destination: destination, Mode: route.Drive, DepartAt: now}
	evening := routing.Query{Origin: origin, Destination: destination, Mode: route.Drive, DepartAt: now.Add(9 * time.Hour)}
	engine.Compute(t.Context(), morning)
	engine.Compute(t.Context(), evening)
	if source.calls != 2 {
		t.Fatal("a rush hour route reused an off-peak cached time")
	}
	if cache.ttl != routing.TrafficCacheTTL {
		t.Fatalf("driving ttl = %v", cache.ttl)
	}
}

func TestMatrixIsCappedAndDegradesSafely(t *testing.T) {
	points := make([]geo.Point, 30)
	for index := range points {
		points[index] = geo.Point{Latitude: 32.6 + float64(index)/1000, Longitude: -16.9}
	}
	source := &provider{matrix: route.Matrix{Legs: []route.Leg{{FromIndex: 0, ToIndex: 1, Evidence: found(route.Walk, 300)}}}}
	matrix := routing.New(source, nil, clock.NewFixed(now)).
		Matrix(t.Context(), routing.MatrixQuery{Points: points, Mode: route.Walk})
	if len(matrix.Points) > routing.MaximumMatrixPoints {
		t.Fatalf("sent %d points, above the cap", len(matrix.Points))
	}
	failing := routing.New(&provider{err: errors.New("quota exhausted")}, nil, clock.NewFixed(now)).
		Matrix(t.Context(), routing.MatrixQuery{Points: points[:3], Mode: route.Walk})
	if len(failing.Legs) != 0 {
		t.Fatal("a failed matrix invented legs")
	}
	tooSmall := routing.New(source, nil, clock.NewFixed(now)).
		Matrix(t.Context(), routing.MatrixQuery{Points: points[:1], Mode: route.Walk})
	if len(tooSmall.Legs) != 0 {
		t.Fatal("a single point produced legs")
	}
}

func TestLeaveByOnlyAnswersWithRealEvidence(t *testing.T) {
	engine := routing.New(&provider{}, nil, clock.NewFixed(now))
	arriveBy := now.Add(3 * time.Hour)
	leaveAt, ok := engine.LeaveBy(arriveBy, found(route.Walk, 1800), 10*time.Minute)
	if !ok || !leaveAt.Equal(arriveBy.Add(-40*time.Minute)) {
		t.Fatalf("leave by = %v %v", leaveAt, ok)
	}
	if _, ok := engine.LeaveBy(arriveBy, route.UnavailableEvidence(origin, destination, route.Walk), 0); ok {
		t.Fatal("a leave-by time was produced without route evidence")
	}
}
