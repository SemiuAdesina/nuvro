package travel

import (
	"context"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/alert"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/evidence"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/itinerary"
	domainlanguage "github.com/SemiuAdesina/nuvro/backend/internal/domain/language"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
)

const OfflinePackVersion = "offline-pack-v1"
const offlinePackLifetime = 24 * time.Hour

type OfflinePack struct {
	Version     string
	TripID      string
	Timezone    string
	TripVersion int
	GeneratedAt time.Time
	ExpiresAt   time.Time
	Days        []itinerary.Day
	Phrases     []domainlanguage.Phrase
	Alerts      []alert.Alert
	WeatherNote string
	Freshness   evidence.Freshness
	Excluded    []string
}

type PhraseSource interface {
	Phrases(ctx context.Context, tripID string) ([]domainlanguage.Phrase, error)
}

func (s *Service) OfflinePack(ctx context.Context, principal user.Principal, tripID string, phrases PhraseSource) (OfflinePack, error) {
	_, place, err := s.context(ctx, principal, tripID)
	if err != nil {
		return OfflinePack{}, err
	}
	now := s.timer.Now().UTC()
	pack := OfflinePack{
		Version: OfflinePackVersion, TripID: tripID, Timezone: place.Timezone,
		GeneratedAt: now, ExpiresAt: now.Add(offlinePackLifetime),
		Excluded:    []string{"provider_photography", "provider_map_tiles", "live_route_times"},
		Freshness:   evidence.Freshness{ObservedAt: now, ValidUntil: now.Add(offlinePackLifetime), Status: evidence.Cached},
		WeatherNote: "cached_at_generation",
	}
	if plan, hasPlan := s.plan(ctx, tripID); hasPlan {
		pack.TripVersion = plan.Version
		pack.Days = plan.Days()
	}
	if phrases != nil {
		if essential, phraseErr := phrases.Phrases(ctx, tripID); phraseErr == nil {
			pack.Phrases = essential
		}
	}
	pack.Alerts = s.visibleAlerts(ctx, tripID, now)
	return pack, nil
}
