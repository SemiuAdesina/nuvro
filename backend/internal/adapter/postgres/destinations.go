package postgres

import (
	"context"

	"github.com/SemiuAdesina/nuvro/backend/internal/apperror"

	queries "github.com/SemiuAdesina/nuvro/backend/internal/adapter/postgres/sqlc"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/destination"
)

const catalogLookupLimit = 500

func (s *Store) Destinations(ctx context.Context, limit int32) ([]destination.Destination, error) {
	rows, err := s.queries(ctx).ListActiveDestinations(ctx, limit)
	if err != nil {
		return nil, databaseError(err)
	}
	catalog := make([]destination.Destination, 0, len(rows))
	for _, row := range rows {
		catalog = append(catalog, destination.Destination{
			ID: row.ID, Name: row.Name, Region: row.Region, CountryCode: row.CountryCode,
			Timezone: row.Timezone, Latitude: row.Latitude, Longitude: row.Longitude,
			PrimaryAirport: row.PrimaryAirport, ClimateTags: tags(row.ClimateTags),
			InterestTags: tags(row.InterestTags), StyleTags: tags(row.StyleTags),
			CrowdProfile: row.CrowdProfile, PaceProfile: row.PaceProfile,
		})
	}
	return catalog, nil
}

func (s *Store) Destination(ctx context.Context, destinationID string) (destination.Destination, error) {
	catalog, err := s.Destinations(ctx, catalogLookupLimit)
	if err != nil {
		return destination.Destination{}, err
	}
	for _, record := range catalog {
		if record.ID == destinationID {
			return record, nil
		}
	}
	return destination.Destination{}, apperror.New(apperror.CodeNotFound, "Destination not found.")
}

func (s *Store) SaveDestination(ctx context.Context, value destination.Destination) error {
	if err := value.Validate(); err != nil {
		return err
	}
	store := s.queries(ctx)
	err := store.UpsertDestination(ctx, queries.UpsertDestinationParams{
		ID: value.ID, Name: value.Name, Region: value.Region, CountryCode: value.CountryCode,
		Timezone: value.Timezone, Latitude: value.Latitude, Longitude: value.Longitude,
		PrimaryAirport: value.PrimaryAirport, CrowdProfile: value.CrowdProfile, PaceProfile: value.PaceProfile,
	})
	if err != nil {
		return databaseError(err)
	}
	for kind, values := range map[string][]string{
		"climate": value.ClimateTags, "interest": value.InterestTags, "style": value.StyleTags,
	} {
		if err := s.replaceTags(ctx, value.ID, kind, values); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) replaceTags(ctx context.Context, destinationID, kind string, values []string) error {
	store := s.queries(ctx)
	if err := store.DeleteDestinationTags(ctx, queries.DeleteDestinationTagsParams{DestinationID: destinationID, Kind: kind}); err != nil {
		return databaseError(err)
	}
	if len(values) == 0 {
		return nil
	}
	return databaseError(store.ReplaceDestinationTags(ctx, queries.ReplaceDestinationTagsParams{
		DestinationID: destinationID, Kind: kind, Column3: values,
	}))
}
