package alerts_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/apperror"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/alert"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/trip"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/id"
	usecase "github.com/SemiuAdesina/nuvro/backend/internal/usecase/alerts"
)

const tripID = "8c2f21b0-4a55-4b39-9f0b-2c7f9c1a1a11"
const owner = "3f1c9c8e-0e3a-4f0a-9b1f-1a2b3c4d5e6f"

var now = time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC)

type trips struct{ err error }

func (t trips) Read(_ context.Context, principal user.Principal, _ string) (trip.Trip, error) {
	if t.err != nil {
		return trip.Trip{}, t.err
	}
	if principal.UserID == "" {
		return trip.Trip{}, apperror.New(apperror.CodeUnauthenticated, "Authentication is required.")
	}
	return trip.Trip{ID: tripID, OwnerID: owner}, nil
}

type store struct {
	byID    map[string]alert.Alert
	saved   []alert.Alert
	expired int
	saveErr error
}

func newStore(values ...alert.Alert) *store {
	byID := map[string]alert.Alert{}
	for _, value := range values {
		byID[value.ID] = value
	}
	return &store{byID: byID}
}

func (s *store) Alerts(context.Context, string) ([]alert.Alert, error) {
	values := make([]alert.Alert, 0, len(s.byID))
	for _, value := range s.byID {
		values = append(values, value)
	}
	return values, nil
}

func (s *store) Alert(_ context.Context, alertID string) (alert.Alert, error) {
	value, found := s.byID[alertID]
	if !found {
		return alert.Alert{}, errors.New("not found")
	}
	return value, nil
}

func (s *store) SaveAlert(_ context.Context, value alert.Alert) (alert.Alert, error) {
	if s.saveErr != nil {
		return alert.Alert{}, s.saveErr
	}
	for index, existing := range s.saved {
		if existing.Fingerprint == value.Fingerprint {
			s.saved[index] = value
			return value, nil
		}
	}
	s.saved = append(s.saved, value)
	s.byID[value.ID] = value
	return value, nil
}

func (s *store) UpdateAlertStatus(_ context.Context, alertID string, status alert.Status) error {
	value, found := s.byID[alertID]
	if !found {
		return errors.New("not found")
	}
	value.Status = status
	s.byID[alertID] = value
	return nil
}

func (s *store) ExpireAlerts(context.Context, string, time.Time) error {
	s.expired++
	return nil
}

func sample(id string, status alert.Status, expiresAt time.Time) alert.Alert {
	return alert.Alert{
		ID: id, TripID: tripID, Type: alert.WeatherImpact, Severity: alert.Attention,
		TitleCode: "alert.weather_impact", Action: alert.AdjustPlan, Status: status,
		Fingerprint: "fingerprint-" + id, CreatedAt: now, EffectiveAt: now.Add(-time.Hour),
		ExpiresAt: expiresAt,
	}
}

func service(t *testing.T, repository *store) *usecase.Service {
	t.Helper()
	return usecase.NewService(trips{}, repository, id.NewUUID(), clock.NewFixed(now))
}

func principal() user.Principal { return user.Principal{UserID: owner} }

func TestListReturnsOnlyVisibleAlertsByDefault(t *testing.T) {
	repository := newStore(
		sample("visible", alert.Unread, now.Add(time.Hour)),
		sample("resolved", alert.Resolved, now.Add(time.Hour)),
	)
	visible, err := service(t, repository).List(t.Context(), principal(), tripID, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(visible) != 1 || visible[0].ID != "visible" {
		t.Fatalf("alerts = %v", visible)
	}
	if repository.expired == 0 {
		t.Fatal("expiry was not applied before listing")
	}
	all, err := service(t, repository).List(t.Context(), principal(), tripID, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 2 {
		t.Fatalf("including resolved returned %d", len(all))
	}
}

func TestStatusTransitionsAreEnforced(t *testing.T) {
	repository := newStore(sample("a", alert.Unread, now.Add(time.Hour)))
	subject := service(t, repository)
	read, err := subject.UpdateStatus(t.Context(), principal(), tripID, "a", alert.Read)
	if err != nil {
		t.Fatal(err)
	}
	if read.Status != alert.Read {
		t.Fatalf("status = %q", read.Status)
	}
	if _, reopenErr := subject.UpdateStatus(t.Context(), principal(), tripID, "a", alert.Unread); apperror.CodeOf(reopenErr) != apperror.CodeInvalidState {
		t.Fatal("an alert was marked unread again")
	}
	resolved, resolveErr := subject.UpdateStatus(t.Context(), principal(), tripID, "a", alert.Resolved)
	if resolveErr != nil {
		t.Fatal(resolveErr)
	}
	if resolved.Status != alert.Resolved {
		t.Fatalf("status = %q", resolved.Status)
	}
	if _, err := subject.UpdateStatus(t.Context(), principal(), tripID, "a", alert.Expired); apperror.CodeOf(err) != apperror.CodeInvalidState {
		t.Fatal("a client expired an alert directly")
	}
}

func TestRepeatingTheSameStatusIsHarmless(t *testing.T) {
	repository := newStore(sample("a", alert.Read, now.Add(time.Hour)))
	value, err := service(t, repository).UpdateStatus(t.Context(), principal(), tripID, "a", alert.Read)
	if err != nil || value.Status != alert.Read {
		t.Fatalf("%v %v", value, err)
	}
}

func TestAlertsFromAnotherTripAreNotFound(t *testing.T) {
	other := sample("a", alert.Unread, now.Add(time.Hour))
	other.TripID = "0c2f21b0-4a55-4b39-9f0b-2c7f9c1a1a99"
	repository := newStore(other)
	if _, err := service(t, repository).UpdateStatus(t.Context(), principal(), tripID, "a", alert.Read); apperror.CodeOf(err) != apperror.CodeNotFound {
		t.Fatal("an alert from another trip was updated")
	}
}

func TestAnonymousCallersAreRefused(t *testing.T) {
	repository := newStore(sample("a", alert.Unread, now.Add(time.Hour)))
	subject := service(t, repository)
	if _, err := subject.List(t.Context(), user.Principal{}, tripID, false); err == nil {
		t.Fatal("an anonymous caller listed alerts")
	}
	if _, err := subject.UpdateStatus(t.Context(), user.Principal{}, tripID, "a", alert.Read); err == nil {
		t.Fatal("an anonymous caller updated an alert")
	}
}

func TestRecordingAssignsIdentifiersAndDeduplicates(t *testing.T) {
	repository := newStore()
	subject := service(t, repository)
	produced := []alert.Alert{
		sample("", alert.Unread, now.Add(time.Hour)),
		sample("", alert.Unread, now.Add(time.Hour)),
	}
	produced[1].Fingerprint = produced[0].Fingerprint
	saved, err := subject.Record(t.Context(), produced)
	if err != nil {
		t.Fatal(err)
	}
	if len(saved) != 2 {
		t.Fatalf("saved %d", len(saved))
	}
	if len(repository.saved) != 1 {
		t.Fatalf("the store kept %d rows for one fingerprint", len(repository.saved))
	}
	for _, value := range saved {
		if value.ID == "" {
			t.Fatal("an alert was recorded without an identifier")
		}
	}
}

func TestStoreFailuresSurface(t *testing.T) {
	repository := newStore()
	repository.saveErr = errors.New("database unavailable")
	if _, err := service(t, repository).Record(t.Context(), []alert.Alert{sample("", alert.Unread, now.Add(time.Hour))}); err == nil {
		t.Fatal("a failing store reported success")
	}
}
