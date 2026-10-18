package destination_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/apperror"
	domain "github.com/SemiuAdesina/nuvro/backend/internal/domain/destination"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
	"github.com/SemiuAdesina/nuvro/backend/internal/engine/intel"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
	usecase "github.com/SemiuAdesina/nuvro/backend/internal/usecase/destination"
)

var now = time.Date(2026, 10, 10, 9, 0, 0, 0, time.UTC)

type catalog struct{ err error }

func (c catalog) Destination(context.Context, string) (domain.Destination, error) {
	if c.err != nil {
		return domain.Destination{}, c.err
	}
	return domain.Destination{
		ID: "madeira-pt", Name: "Madeira", CountryCode: "PT", Timezone: "Atlantic/Madeira",
		Latitude: 32.7607, Longitude: -16.9595, PrimaryAirport: "FNC",
	}, nil
}

type profiles struct{ interests []string }

func (p profiles) Preferences(context.Context, string) (user.Preferences, error) {
	return user.Preferences{Interests: p.interests}, nil
}

func (p profiles) Profile(context.Context, string) (user.Profile, error) {
	return user.Profile{}, nil
}

type describer struct{ seen intel.Request }

func (d *describer) Describe(_ context.Context, request intel.Request) intel.Intelligence {
	d.seen = request
	return intel.Intelligence{Destination: request.Destination, GeneratedAt: now}
}

func TestDetailPassesTravellerInterestsAndDestination(t *testing.T) {
	engine := &describer{}
	service := usecase.NewService(catalog{}, profiles{interests: []string{"food", "nature"}}, engine, clock.NewFixed(now))
	result, err := service.Detail(t.Context(), user.Principal{UserID: "traveller"}, usecase.Request{
		DestinationID: "madeira-pt", Language: "en",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Destination.ID != "madeira-pt" {
		t.Fatalf("destination = %+v", result.Destination)
	}
	if len(engine.seen.Interests) != 2 {
		t.Fatalf("interests were not passed: %v", engine.seen.Interests)
	}
	if engine.seen.Language != "en" {
		t.Fatal("language was not passed")
	}
}

func TestAbsentDatesFallBackToANearTermWindow(t *testing.T) {
	engine := &describer{}
	service := usecase.NewService(catalog{}, profiles{}, engine, clock.NewFixed(now))
	if _, err := service.Detail(t.Context(), user.Principal{UserID: "traveller"}, usecase.Request{
		DestinationID: "madeira-pt",
	}); err != nil {
		t.Fatal(err)
	}
	if engine.seen.Start.IsZero() || !engine.seen.End.After(engine.seen.Start) {
		t.Fatalf("window = %v to %v", engine.seen.Start, engine.seen.End)
	}
	if engine.seen.Start.Before(now.Truncate(24 * time.Hour)) {
		t.Fatal("the fallback window starts in the past")
	}
}

func TestSuppliedDatesAreUsedWhenSensible(t *testing.T) {
	engine := &describer{}
	service := usecase.NewService(catalog{}, profiles{}, engine, clock.NewFixed(now))
	start := now.AddDate(0, 0, 5)
	end := now.AddDate(0, 0, 12)
	if _, err := service.Detail(t.Context(), user.Principal{UserID: "traveller"}, usecase.Request{
		DestinationID: "madeira-pt", Start: start, End: end,
	}); err != nil {
		t.Fatal(err)
	}
	if !engine.seen.Start.Equal(start) || !engine.seen.End.Equal(end) {
		t.Fatalf("window = %v to %v", engine.seen.Start, engine.seen.End)
	}
	reversed := usecase.Request{DestinationID: "madeira-pt", Start: end, End: start}
	if _, err := service.Detail(t.Context(), user.Principal{UserID: "traveller"}, reversed); err != nil {
		t.Fatal(err)
	}
	if engine.seen.End.Before(engine.seen.Start) {
		t.Fatal("a reversed range was passed through")
	}
}

func TestUnknownDestinationAndAnonymousCallersAreRefused(t *testing.T) {
	service := usecase.NewService(catalog{err: apperror.New(apperror.CodeNotFound, "Destination not found.")},
		profiles{}, &describer{}, clock.NewFixed(now))
	if _, err := service.Detail(t.Context(), user.Principal{UserID: "traveller"}, usecase.Request{
		DestinationID: "atlantis",
	}); apperror.CodeOf(err) != apperror.CodeNotFound {
		t.Fatal("an unknown destination was described")
	}
	open := usecase.NewService(catalog{}, profiles{}, &describer{}, clock.NewFixed(now))
	if _, err := open.Detail(t.Context(), user.Principal{}, usecase.Request{DestinationID: "madeira-pt"}); apperror.CodeOf(err) != apperror.CodeUnauthenticated {
		t.Fatal("an anonymous caller received destination intelligence")
	}
}

func TestFailingPreferencesDoNotBlockTheDestination(t *testing.T) {
	engine := &describer{}
	service := usecase.NewService(catalog{}, failingProfiles{}, engine, clock.NewFixed(now))
	if _, err := service.Detail(t.Context(), user.Principal{UserID: "traveller"}, usecase.Request{
		DestinationID: "madeira-pt",
	}); err != nil {
		t.Fatalf("a failing preference lookup broke the page: %v", err)
	}
	if len(engine.seen.Interests) != 0 {
		t.Fatal("interests were invented")
	}
}

type failingProfiles struct{}

func (failingProfiles) Preferences(context.Context, string) (user.Preferences, error) {
	return user.Preferences{}, errors.New("preferences unavailable")
}

func (failingProfiles) Profile(context.Context, string) (user.Profile, error) {
	return user.Profile{}, errors.New("profile unavailable")
}
