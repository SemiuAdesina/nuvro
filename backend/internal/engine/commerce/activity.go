package commerce

import (
	"slices"
	"time"

	domain "github.com/SemiuAdesina/nuvro/backend/internal/domain/commerce"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/geo"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/weather"
)

type ActivityContext struct {
	Now                   time.Time
	Centre                geo.Point
	BudgetMinor           int64
	FreeFrom              time.Time
	FreeUntil             time.Time
	Weather               weather.Window
	MaximumDistanceMeters int
}

func RankActivities(offers []domain.ActivityOffer, context ActivityContext) []domain.ActivityOffer {
	ranked := acceptActivities(offers, context)
	slices.SortStableFunc(ranked, func(left, right domain.ActivityOffer) int {
		if order := activityScore(right, context) - activityScore(left, context); order != 0 {
			return order
		}
		if order := int(left.Price.Minor() - right.Price.Minor()); order != 0 {
			return order
		}
		return slices.Compare([]string{left.ID}, []string{right.ID})
	})
	for index := range ranked {
		ranked[index].Rank = index + 1
		ranked[index].Reasons = activityReasons(ranked[index], ranked, context)
	}
	return ranked
}

func acceptActivities(offers []domain.ActivityOffer, context ActivityContext) []domain.ActivityOffer {
	accepted := make([]domain.ActivityOffer, 0, len(offers))
	seen := make(map[string]struct{}, len(offers))
	for _, offer := range offers {
		if offer.Validate() != nil || !offer.Available || !offer.FreshnessAt(context.Now).Bookable() {
			continue
		}
		if _, duplicate := seen[offer.Signature()]; duplicate {
			continue
		}
		seen[offer.Signature()] = struct{}{}
		accepted = append(accepted, offer)
	}
	return accepted
}

func activityScore(offer domain.ActivityOffer, context ActivityContext) int {
	score := 60
	if fitsWindow(offer, context) {
		score += 25
	} else if !context.FreeFrom.IsZero() {
		score -= 40
	}
	if !context.Centre.IsZero() {
		metres := int(context.Centre.DistanceMeters(offer.Location))
		score += proximityScore(metres)
		if context.MaximumDistanceMeters > 0 && metres > context.MaximumDistanceMeters {
			score -= 30
		}
	}
	if context.BudgetMinor > 0 {
		if offer.Price.Minor() <= context.BudgetMinor {
			score += 12
		} else {
			score -= 25
		}
	}
	if impact := weatherImpact(offer, context); impact == weather.HighImpact {
		score -= 45
	} else if impact == weather.ModerateImpact {
		score -= 10
	}
	score += int(offer.Rating * 3)
	return score
}

func fitsWindow(offer domain.ActivityOffer, context ActivityContext) bool {
	if context.FreeFrom.IsZero() || context.FreeUntil.IsZero() {
		return true
	}
	if offer.StartsAt.IsZero() {
		return offer.DurationMinutes <= int(context.FreeUntil.Sub(context.FreeFrom).Minutes())
	}
	ends := offer.StartsAt.Add(time.Duration(offer.DurationMinutes) * time.Minute)
	return !offer.StartsAt.Before(context.FreeFrom) && !ends.After(context.FreeUntil)
}

func weatherImpact(offer domain.ActivityOffer, context ActivityContext) weather.Impact {
	if !context.Weather.Known() || !offer.Outdoor {
		return weather.UnknownImpact
	}
	moment := offer.StartsAt
	if moment.IsZero() {
		moment = context.FreeFrom
	}
	if moment.IsZero() {
		return weather.UnknownImpact
	}
	observation, found := context.Weather.HourAt(moment)
	if !found {
		return weather.UnknownImpact
	}
	return weather.Assess(observation, weather.Outdoor)
}

func activityReasons(offer domain.ActivityOffer, all []domain.ActivityOffer, context ActivityContext) []domain.Reason {
	reasons := make([]domain.Reason, 0, 4)
	if cheapestActivity(offer, all) {
		reasons = append(reasons, domain.BestPrice)
	}
	if fitsWindow(offer, context) && !context.FreeFrom.IsZero() {
		reasons = append(reasons, domain.FitsItinerary)
	}
	if context.BudgetMinor > 0 && offer.Price.Minor() <= context.BudgetMinor {
		reasons = append(reasons, domain.WithinBudget)
	}
	if offer.Rating >= 4.5 {
		reasons = append(reasons, domain.HighlyRated)
	}
	return reasons
}

func cheapestActivity(offer domain.ActivityOffer, all []domain.ActivityOffer) bool {
	for _, candidate := range all {
		if candidate.Price.Minor() < offer.Price.Minor() {
			return false
		}
	}
	return true
}

func RankTransfers(offers []domain.TransferOffer, now time.Time) []domain.TransferOffer {
	accepted := make([]domain.TransferOffer, 0, len(offers))
	seen := make(map[string]struct{}, len(offers))
	for _, offer := range offers {
		if offer.Validate() != nil || !offer.Available || !offer.FreshnessAt(now).Bookable() {
			continue
		}
		if _, duplicate := seen[offer.Signature()]; duplicate {
			continue
		}
		seen[offer.Signature()] = struct{}{}
		accepted = append(accepted, offer)
	}
	slices.SortStableFunc(accepted, func(left, right domain.TransferOffer) int {
		if order := left.DurationMinutes - right.DurationMinutes; order != 0 {
			return order
		}
		if order := int(left.Price.Minor() - right.Price.Minor()); order != 0 {
			return order
		}
		return slices.Compare([]string{left.ID}, []string{right.ID})
	})
	for index := range accepted {
		accepted[index].Rank = index + 1
		if index == 0 {
			accepted[index].Reasons = []domain.Reason{domain.ShortestTrip}
		}
	}
	return accepted
}
