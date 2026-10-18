package journey

import (
	"context"

	"github.com/SemiuAdesina/nuvro/backend/internal/apperror"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/geo"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/weather"
)

type destinationRecord struct {
	ID        string
	Latitude  float64
	Longitude float64
	Timezone  string
	Country   string
}

func (s *Service) context(ctx context.Context, principal user.Principal, tripID string) (tripContext, error) {
	record, err := s.trips.Read(ctx, principal, tripID)
	if err != nil {
		return tripContext{}, err
	}
	if record.DestinationID == nil || record.StartDate == nil || record.EndDate == nil {
		return tripContext{}, apperror.Wrap(apperror.CodeInvalidState,
			"This trip needs a destination and travel dates first.", errNoDestination)
	}
	role, err := s.trips.Role(ctx, tripID, principal.UserID)
	if err != nil {
		return tripContext{}, err
	}
	place, err := s.catalog.Destination(ctx, *record.DestinationID)
	if err != nil {
		return tripContext{}, apperror.New(apperror.CodeNotFound, "Destination not found.")
	}
	return tripContext{
		Trip: record, Role: role,
		Destination: destinationRecord{
			ID: place.ID, Latitude: place.Latitude, Longitude: place.Longitude,
			Timezone: place.Timezone, Country: place.CountryCode,
		},
	}, nil
}

func (s *Service) forecast(ctx context.Context, point geo.Point, context tripContext, language string) weather.Window {
	if s.forecaster == nil {
		return weather.Unavailable(weather.ProviderUnavailable)
	}
	return s.forecaster.For(ctx, point, *context.Trip.StartDate, *context.Trip.EndDate, language)
}
