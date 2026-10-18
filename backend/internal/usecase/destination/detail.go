package destination

import (
	"context"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/apperror"
	domain "github.com/SemiuAdesina/nuvro/backend/internal/domain/destination"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
	"github.com/SemiuAdesina/nuvro/backend/internal/engine/intel"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
)

type Catalog interface {
	Destination(context.Context, string) (domain.Destination, error)
}

type Preferences interface {
	Preferences(context.Context, string) (user.Preferences, error)
	Profile(context.Context, string) (user.Profile, error)
}

type Describer interface {
	Describe(context.Context, intel.Request) intel.Intelligence
}

type Service struct {
	catalog     Catalog
	preferences Preferences
	describer   Describer
	timer       clock.Clock
}

func NewService(catalog Catalog, preferences Preferences, describer Describer, timer clock.Clock) *Service {
	return &Service{catalog: catalog, preferences: preferences, describer: describer, timer: timer}
}

type Request struct {
	DestinationID string
	Start         time.Time
	End           time.Time
	Language      string
}

func (s *Service) Detail(ctx context.Context, principal user.Principal, request Request) (intel.Intelligence, error) {
	if principal.UserID == "" {
		return intel.Intelligence{}, apperror.New(apperror.CodeUnauthenticated, "Authentication is required.")
	}
	record, err := s.catalog.Destination(ctx, request.DestinationID)
	if err != nil {
		return intel.Intelligence{}, apperror.Wrap(apperror.CodeNotFound, "Destination not found.", err)
	}
	start, end := s.window(request)
	return s.describer.Describe(ctx, intel.Request{
		Destination: record, Start: start, End: end,
		Language: request.Language, Interests: s.interests(ctx, principal),
	}), nil
}

func (s *Service) window(request Request) (time.Time, time.Time) {
	if !request.Start.IsZero() && !request.End.IsZero() && !request.End.Before(request.Start) {
		return request.Start.UTC(), request.End.UTC()
	}
	now := s.timer.Now().UTC().Truncate(24 * time.Hour)
	return now, now.AddDate(0, 0, 3)
}

func (s *Service) interests(ctx context.Context, principal user.Principal) []string {
	if s.preferences == nil {
		return nil
	}
	preferences, err := s.preferences.Preferences(ctx, principal.UserID)
	if err != nil {
		return nil
	}
	return preferences.Interests
}
