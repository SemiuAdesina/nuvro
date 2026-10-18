package places

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/place"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
)

const NearbyCacheTTL = 15 * time.Minute
const DetailsCacheTTL = 6 * time.Hour
const maximumResults = 40

type Engine struct {
	search  SearchProvider
	details DetailsProvider
	cache   Cache
	timer   clock.Clock
}

func New(search SearchProvider, details DetailsProvider, cache Cache, timer clock.Clock) *Engine {
	return &Engine{search: search, details: details, cache: cache, timer: timer}
}

func (e *Engine) Configured() bool { return e.search != nil }

func (e *Engine) Nearby(ctx context.Context, query NearbyQuery) ([]place.Place, error) {
	if e.search == nil {
		return nil, ErrUnavailable
	}
	query.Limit = boundedLimit(query.Limit)
	key := nearbyKey(query)
	if cached, found := e.fromCache(ctx, key); found {
		return cached, nil
	}
	found, err := e.search.Nearby(ctx, query)
	if err != nil {
		return nil, err
	}
	results := accept(found, query.Limit)
	e.store(ctx, key, results, NearbyCacheTTL)
	return results, nil
}

func (e *Engine) Text(ctx context.Context, query TextQuery) ([]place.Place, error) {
	if e.search == nil {
		return nil, ErrUnavailable
	}
	query.Limit = boundedLimit(query.Limit)
	if strings.TrimSpace(query.Text) == "" {
		return nil, ErrUnavailable
	}
	found, err := e.search.Text(ctx, query)
	if err != nil {
		return nil, err
	}
	return accept(found, query.Limit), nil
}

func (e *Engine) Resolve(ctx context.Context, name string, query TextQuery) (place.Place, bool) {
	query.Text = name
	query.Limit = 5
	candidates, err := e.Text(ctx, query)
	if err != nil || len(candidates) == 0 {
		return place.Place{}, false
	}
	wanted := normalise(name)
	for _, candidate := range candidates {
		if normalise(candidate.Name) == wanted {
			return candidate, true
		}
	}
	return place.Place{}, false
}

func (e *Engine) Details(ctx context.Context, query DetailsQuery) (place.Place, error) {
	if e.details == nil {
		return place.Place{}, ErrUnavailable
	}
	key := detailsKey(query)
	if cached, found := e.fromCache(ctx, key); found && len(cached) == 1 {
		return cached[0], nil
	}
	found, err := e.details.Details(ctx, query)
	if err != nil {
		return place.Place{}, err
	}
	if err := found.Validate(); err != nil {
		return place.Place{}, err
	}
	e.store(ctx, key, []place.Place{found}, DetailsCacheTTL)
	return found, nil
}

func (e *Engine) fromCache(ctx context.Context, key string) ([]place.Place, bool) {
	if e.cache == nil {
		return nil, false
	}
	return e.cache.Places(ctx, key)
}

func (e *Engine) store(ctx context.Context, key string, values []place.Place, ttl time.Duration) {
	if e.cache == nil || len(values) == 0 {
		return
	}
	e.cache.StorePlaces(ctx, key, values, ttl)
}

func accept(found []place.Place, limit int) []place.Place {
	results := make([]place.Place, 0, len(found))
	seen := make(map[string]struct{}, len(found))
	for _, candidate := range found {
		if err := candidate.Validate(); err != nil {
			continue
		}
		if _, duplicate := seen[candidate.ID]; duplicate {
			continue
		}
		seen[candidate.ID] = struct{}{}
		results = append(results, candidate)
		if len(results) == limit {
			break
		}
	}
	return results
}

func boundedLimit(limit int) int {
	if limit < 1 || limit > maximumResults {
		return maximumResults
	}
	return limit
}

func nearbyKey(query NearbyQuery) string {
	parts := []string{
		coordinate(query.Centre.Latitude), coordinate(query.Centre.Longitude),
		strconv.Itoa(query.RadiusMeters), strconv.FormatBool(query.OpenNow), query.Language,
	}
	for _, category := range query.Categories {
		parts = append(parts, string(category))
	}
	return strings.Join(parts, "_")
}

func detailsKey(query DetailsQuery) string {
	return query.Reference.Provider + "_" + query.Reference.ExternalID + "_" + query.Language
}

func coordinate(value float64) string {
	return strconv.FormatFloat(value, 'f', 3, 64)
}

func normalise(value string) string {
	return strings.Join(strings.Fields(strings.ToLower(strings.TrimSpace(value))), " ")
}
