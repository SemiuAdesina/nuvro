package googleplaces

import (
	"strings"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/destination"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/evidence"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/geo"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/place"
)

const Provider = "google_places"

func toPlace(payload placePayload, observedAt time.Time) (place.Place, bool) {
	if payload.ID == "" || payload.DisplayName.Text == "" {
		return place.Place{}, false
	}
	result := place.Place{
		ID:          Provider + ":" + payload.ID,
		References:  []place.Reference{{Provider: Provider, ExternalID: payload.ID}},
		Name:        payload.DisplayName.Text,
		Category:    category(payload.PrimaryType, payload.Types),
		RawTypes:    payload.Types,
		Location:    geo.Point{Latitude: payload.Location.Latitude, Longitude: payload.Location.Longitude},
		Address:     payload.FormattedAddress,
		Rating:      clampRating(payload.Rating),
		ReviewCount: max(payload.UserRatingCount, 0),
		PriceLevel:  priceLevel(payload.PriceLevel),
		Status:      openingStatus(payload),
		Hours:       hours(payload, observedAt),
		Media:       mediaAssets(payload.Photos),
		Attribution: attribution(payload.Photos),
		Freshness: evidence.Freshness{
			ObservedAt: observedAt.UTC(), ValidUntil: observedAt.UTC().Add(15 * time.Minute), Status: evidence.Live,
		},
	}
	if result.Validate() != nil {
		return place.Place{}, false
	}
	return result, true
}

func clampRating(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 5 {
		return 5
	}
	return value
}

func priceLevel(raw string) int {
	switch raw {
	case "PRICE_LEVEL_FREE", "PRICE_LEVEL_INEXPENSIVE":
		return 1
	case "PRICE_LEVEL_MODERATE":
		return 2
	case "PRICE_LEVEL_EXPENSIVE":
		return 3
	case "PRICE_LEVEL_VERY_EXPENSIVE":
		return 4
	default:
		return 0
	}
}

func openingStatus(payload placePayload) place.OpeningStatus {
	for _, source := range []*openingHours{payload.CurrentOpeningHours, payload.RegularOpeningHours} {
		if source == nil || source.OpenNow == nil {
			continue
		}
		if *source.OpenNow {
			return place.Open
		}
		return place.Closed
	}
	return place.UnknownStatus
}

func hours(payload placePayload, observedAt time.Time) place.OpeningHours {
	source := payload.RegularOpeningHours
	if source == nil || len(source.Periods) == 0 {
		return place.OpeningHours{ObservedAt: observedAt.UTC()}
	}
	intervals := make([]place.Interval, 0, len(source.Periods))
	for _, entry := range source.Periods {
		interval := place.Interval{
			Weekday: time.Weekday(entry.Open.Day % 7),
			Opens:   entry.Open.Hour*60 + entry.Open.Minute,
			Closes:  entry.Close.Hour*60 + entry.Close.Minute,
		}
		if interval.Closes == 0 {
			interval.Closes = 24 * 60
		}
		if !interval.Valid() {
			continue
		}
		intervals = append(intervals, interval)
	}
	if len(intervals) == 0 {
		return place.OpeningHours{ObservedAt: observedAt.UTC()}
	}
	return place.OpeningHours{Intervals: intervals, ObservedAt: observedAt.UTC(), Known: true}
}

func mediaAssets(photos []photo) []destination.Media {
	assets := make([]destination.Media, 0, len(photos))
	for _, entry := range photos {
		if entry.Name == "" || entry.WidthPx <= 0 || entry.HeightPx <= 0 {
			continue
		}
		assets = append(assets, destination.Media{
			URL:         photoURL(entry.Name, entry.WidthPx),
			Width:       entry.WidthPx,
			Height:      entry.HeightPx,
			AspectRatio: float64(entry.WidthPx) / float64(entry.HeightPx),
			Attribution: photoAttribution(entry),
		})
	}
	return assets
}

func photoAttribution(entry photo) []evidence.Attribution {
	attributions := make([]evidence.Attribution, 0, len(entry.AuthorAttributions))
	for _, author := range entry.AuthorAttributions {
		if author.DisplayName == "" {
			continue
		}
		attributions = append(attributions, evidence.Attribution{
			Text: author.DisplayName, URI: author.URI, Provider: Provider,
		})
	}
	return attributions
}

func attribution(photos []photo) []evidence.Attribution {
	seen := map[string]struct{}{}
	attributions := make([]evidence.Attribution, 0)
	for _, entry := range photos {
		for _, value := range photoAttribution(entry) {
			if _, duplicate := seen[value.Text]; duplicate {
				continue
			}
			seen[value.Text] = struct{}{}
			attributions = append(attributions, value)
		}
	}
	return attributions
}

func photoURL(name string, width int) string {
	return "https://places.googleapis.com/v1/" + strings.TrimPrefix(name, "/") + "/media?maxWidthPx=" + itoa(width)
}

func itoa(value int) string {
	if value <= 0 {
		return "0"
	}
	digits := ""
	for value > 0 {
		digits = string(rune('0'+value%10)) + digits
		value /= 10
	}
	return digits
}
