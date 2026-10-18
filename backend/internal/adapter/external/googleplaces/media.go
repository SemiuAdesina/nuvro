package googleplaces

import (
	"context"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/evidence"
	"github.com/SemiuAdesina/nuvro/backend/internal/engine/media"
	"github.com/SemiuAdesina/nuvro/backend/internal/engine/places"
)

type MediaAdapter struct {
	search   places.SearchProvider
	resolver DestinationResolver
}

type DestinationResolver interface {
	Query(ctx context.Context, destinationID string) (places.TextQuery, bool)
}

func NewMediaAdapter(search places.SearchProvider, resolver DestinationResolver) *MediaAdapter {
	return &MediaAdapter{search: search, resolver: resolver}
}

func (m *MediaAdapter) Candidates(ctx context.Context, request media.Context) ([]media.Candidate, error) {
	if m.search == nil || m.resolver == nil {
		return nil, nil
	}
	query, found := m.resolver.Query(ctx, request.DestinationID)
	if !found {
		return nil, nil
	}
	query.Language = request.Language
	query.Limit = 5
	results, err := m.search.Text(ctx, query)
	if err != nil {
		return nil, err
	}
	candidates := make([]media.Candidate, 0, len(results))
	for _, result := range results {
		for _, asset := range result.Media {
			candidates = append(candidates, media.Candidate{
				Media: asset, Source: evidence.Source{Provider: Provider, Freshness: result.Freshness}, ProviderReference: result.ID,
			})
		}
	}
	return candidates, nil
}

var _ media.CandidateReader = (*MediaAdapter)(nil)
