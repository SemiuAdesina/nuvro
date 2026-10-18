package commerce

import (
	"slices"
	"time"

	domain "github.com/SemiuAdesina/nuvro/backend/internal/domain/commerce"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/geo"
)

const FlightRankingVersion = "flight-rank-v1"
const StayRankingVersion = "stay-rank-v1"
const ActivityRankingVersion = "activity-rank-v1"
const TransferRankingVersion = "transfer-rank-v1"

type FlightContext struct {
	Now              time.Time
	PreferFewerStops bool
	MaximumMinutes   int
	BudgetMinor      int64
}

func RankFlights(offers []domain.FlightOffer, context FlightContext) []domain.FlightOffer {
	ranked := acceptFlights(offers, context.Now)
	slices.SortStableFunc(ranked, func(left, right domain.FlightOffer) int {
		if order := flightScore(right, context) - flightScore(left, context); order != 0 {
			return order
		}
		if order := int(left.Price.Minor() - right.Price.Minor()); order != 0 {
			return order
		}
		return slices.Compare([]string{left.ID}, []string{right.ID})
	})
	for index := range ranked {
		ranked[index].Rank = index + 1
		ranked[index].Reasons = flightReasons(ranked[index], ranked, context)
	}
	return ranked
}

func acceptFlights(offers []domain.FlightOffer, now time.Time) []domain.FlightOffer {
	accepted := make([]domain.FlightOffer, 0, len(offers))
	seen := make(map[string]struct{}, len(offers))
	for _, offer := range offers {
		if offer.Validate() != nil || !offer.Available || !offer.FreshnessAt(now).Bookable() {
			continue
		}
		signature := offer.Signature()
		if existing, duplicate := seen[signature]; duplicate {
			_ = existing
			continue
		}
		seen[signature] = struct{}{}
		accepted = append(accepted, offer)
	}
	return accepted
}

func flightScore(offer domain.FlightOffer, context FlightContext) int {
	score := 100
	score -= min(offer.DurationMinutes/30, 40)
	score -= min(offer.Stops*12, 30)
	score -= min(offer.LayoverMinutes/45, 15)
	if context.PreferFewerStops && offer.Stops == 0 {
		score += 10
	}
	if context.MaximumMinutes > 0 && offer.DurationMinutes > context.MaximumMinutes {
		score -= 30
	}
	if context.BudgetMinor > 0 && offer.Price.Minor() > context.BudgetMinor {
		score -= 25
	}
	if offer.BaggageIncluded {
		score += 5
	}
	if offer.Changeable {
		score += 3
	}
	return score
}

func flightReasons(offer domain.FlightOffer, all []domain.FlightOffer, context FlightContext) []domain.Reason {
	reasons := make([]domain.Reason, 0, 4)
	if isCheapest(offer, all) {
		reasons = append(reasons, domain.BestPrice)
	}
	if isShortest(offer, all) {
		reasons = append(reasons, domain.ShortestTrip)
	}
	if offer.Stops == 0 {
		reasons = append(reasons, domain.FewerStops)
	}
	if context.BudgetMinor > 0 && offer.Price.Minor() <= context.BudgetMinor {
		reasons = append(reasons, domain.WithinBudget)
	}
	if offer.Changeable {
		reasons = append(reasons, domain.FlexibleCancellation)
	}
	return reasons
}

func isCheapest(offer domain.FlightOffer, all []domain.FlightOffer) bool {
	for _, candidate := range all {
		if candidate.Price.Minor() < offer.Price.Minor() {
			return false
		}
	}
	return true
}

func isShortest(offer domain.FlightOffer, all []domain.FlightOffer) bool {
	for _, candidate := range all {
		if candidate.DurationMinutes < offer.DurationMinutes {
			return false
		}
	}
	return true
}

type StayContext struct {
	Now               time.Time
	ItineraryCentre   geo.Point
	BudgetMinor       int64
	MaximumWalkMeters int
}

func RankStays(offers []domain.StayOffer, context StayContext) []domain.StayOffer {
	ranked := acceptStays(offers, context.Now)
	slices.SortStableFunc(ranked, func(left, right domain.StayOffer) int {
		if order := stayScore(right, context) - stayScore(left, context); order != 0 {
			return order
		}
		if order := int(left.Price.Minor() - right.Price.Minor()); order != 0 {
			return order
		}
		return slices.Compare([]string{left.ID}, []string{right.ID})
	})
	for index := range ranked {
		ranked[index].Rank = index + 1
		ranked[index].Reasons = stayReasons(ranked[index], ranked, context)
	}
	return ranked
}

func acceptStays(offers []domain.StayOffer, now time.Time) []domain.StayOffer {
	accepted := make([]domain.StayOffer, 0, len(offers))
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
	return accepted
}

func stayScore(offer domain.StayOffer, context StayContext) int {
	score := 60
	if !context.ItineraryCentre.IsZero() {
		metres := int(context.ItineraryCentre.DistanceMeters(offer.Location))
		score += proximityScore(metres)
		if context.MaximumWalkMeters > 0 && metres > context.MaximumWalkMeters {
			score -= 20
		}
	}
	if context.BudgetMinor > 0 {
		if offer.Price.Minor() <= context.BudgetMinor {
			score += 15
		} else {
			score -= 25
		}
	}
	if offer.ReviewCount >= 50 {
		score += int(offer.Rating * 4)
	}
	if offer.FreeCancellation {
		score += 6
	}
	return score
}

func proximityScore(metres int) int {
	switch {
	case metres <= 1000:
		return 30
	case metres <= 3000:
		return 20
	case metres <= 8000:
		return 8
	default:
		return max(4-metres/5000, -25)
	}
}

func stayReasons(offer domain.StayOffer, all []domain.StayOffer, context StayContext) []domain.Reason {
	reasons := make([]domain.Reason, 0, 4)
	if cheapestStay(offer, all) {
		reasons = append(reasons, domain.BestPrice)
	}
	if !context.ItineraryCentre.IsZero() && nearestStay(offer, all, context.ItineraryCentre) {
		reasons = append(reasons, domain.BestLocation)
	}
	if context.BudgetMinor > 0 && offer.Price.Minor() <= context.BudgetMinor {
		reasons = append(reasons, domain.WithinBudget)
	}
	if offer.FreeCancellation {
		reasons = append(reasons, domain.FlexibleCancellation)
	}
	if offer.ReviewCount >= 50 && offer.Rating >= 4.3 {
		reasons = append(reasons, domain.HighlyRated)
	}
	return reasons
}

func cheapestStay(offer domain.StayOffer, all []domain.StayOffer) bool {
	for _, candidate := range all {
		if candidate.Price.Minor() < offer.Price.Minor() {
			return false
		}
	}
	return true
}

func nearestStay(offer domain.StayOffer, all []domain.StayOffer, centre geo.Point) bool {
	distance := centre.DistanceMeters(offer.Location)
	for _, candidate := range all {
		if centre.DistanceMeters(candidate.Location) < distance {
			return false
		}
	}
	return true
}
