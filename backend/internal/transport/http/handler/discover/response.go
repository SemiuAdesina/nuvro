package discover

import (
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/destination"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/feasibility"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/money"
	"github.com/SemiuAdesina/nuvro/backend/internal/engine/recommendation"
	"github.com/SemiuAdesina/nuvro/backend/internal/transport/http/api"
	"github.com/SemiuAdesina/nuvro/backend/internal/usecase/discover"
)

func response(result discover.Response) api.MatchResponse {
	recommendations := make([]api.Recommendation, 0, len(result.Result.Recommendations))
	for _, value := range result.Result.Recommendations {
		recommendations = append(recommendations, recommendationOf(value))
	}
	missing := make([]api.MatchResponseMissingContext, 0, len(result.MissingContext))
	for _, field := range result.MissingContext {
		missing = append(missing, api.MatchResponseMissingContext(field))
	}
	return api.MatchResponse{
		GeneratedAt: result.GeneratedAt, AlgorithmVersion: result.EvidenceVersion,
		RankingVersion: result.Result.RankingVersion, MissingContext: missing,
		Recommendations: recommendations,
	}
}

func recommendationOf(value recommendation.Recommendation) api.Recommendation {
	return api.Recommendation{
		Rank:          value.Rank,
		Destination:   destinationOf(value.Destination),
		PersonalMatch: matchOf(value.Match),
		Access:        accessOf(value.Feasibility.Access),
		Affordability: affordabilityOf(value.Feasibility.Affordability),
		Flight:        flightOf(value.Feasibility.Flight),
		Reasons:       stringsOrEmpty(value.Reasons),
		Cautions:      stringsOrEmpty(value.Feasibility.Cautions),
		Summary:       value.Summary,
	}
}

func destinationOf(value destination.Destination) api.DestinationSummary {
	summary := api.DestinationSummary{
		Id: value.ID, Name: value.Name, CountryCode: value.CountryCode, Timezone: value.Timezone,
		PrimaryAirport: value.PrimaryAirport,
		Coordinates:    api.GeoPoint{Latitude: float32(value.Latitude), Longitude: float32(value.Longitude)},
	}
	if value.Region != "" {
		region := value.Region
		summary.Region = &region
	}
	return summary
}

func matchOf(score destination.MatchScore) api.PersonalMatch {
	components := make([]api.MatchComponent, 0, len(score.Components))
	for _, component := range score.Components {
		components = append(components, api.MatchComponent{
			Key: component.Key, Label: component.Label, AwardedPoints: component.AwardedPoints,
			PossiblePoints: component.PossiblePoints, Explanation: component.Explanation,
			Assessment: api.MatchComponentAssessment(component.Assessment),
		})
	}
	return api.PersonalMatch{Score: score.Score, AlgorithmVersion: score.WeightsVersion, Components: components}
}

func accessOf(access feasibility.Access) api.Access {
	value := api.Access{
		Status: api.AccessStatus(access.Status), EntryMode: api.AccessEntryMode(access.EntryMode),
		Preparation: api.AccessPreparation(access.Preparation), TransitRisk: api.AccessTransitRisk(access.TransitRisk),
		Confidence: api.Confidence(access.Confidence), Reasons: stringsOrEmpty(access.Reasons),
	}
	if access.SourceLabel != "" {
		label := access.SourceLabel
		value.SourceLabel = &label
	}
	value.ObservedAt = instant(access.ObservedAt)
	return value
}

func affordabilityOf(affordability feasibility.Affordability) api.Affordability {
	value := api.Affordability{
		Status: api.AffordabilityStatus(affordability.Status), Budget: moneyOf(affordability.Budget),
		Confidence: api.Confidence(affordability.Confidence), ObservedAt: instant(affordability.ObservedAt),
	}
	if affordability.HasEstimate {
		value.EstimatedTrip = &api.CostRange{
			Minimum: moneyOf(affordability.Estimate.Minimum), Expected: moneyOf(affordability.Estimate.Expected),
			Maximum: moneyOf(affordability.Estimate.Maximum),
		}
	}
	if affordability.Conversion != nil {
		value.Conversion = &api.Conversion{
			Original: moneyOf(affordability.Conversion.Source), Reference: moneyOf(affordability.Conversion.Target),
			RateMicros: affordability.Conversion.Rate.Micros(), Source: affordability.Conversion.Provider,
			ObservedAt: affordability.Conversion.ObservedAt,
		}
	}
	return value
}

func flightOf(flight feasibility.Flight) api.FlightSummary {
	value := api.FlightSummary{Status: api.FlightSummaryStatus(flight.Status), Confidence: api.Confidence(flight.Confidence)}
	if flight.Status == feasibility.FlightAvailable {
		stops := flight.Stops
		duration := flight.DurationMinutes
		value.Stops = &stops
		value.DurationMinutes = &duration
		if flight.HasPrice {
			value.EstimatedPrice = &api.Money{AmountMinor: flight.PriceMinor, Currency: flight.PriceCurrency}
		}
	}
	value.ObservedAt = instant(flight.ObservedAt)
	return value
}

func moneyOf(amount money.Amount) api.Money {
	return api.Money{AmountMinor: amount.Minor(), Currency: amount.Currency().String()}
}

func instant(value time.Time) *time.Time {
	if value.IsZero() {
		return nil
	}
	moment := value.UTC()
	return &moment
}

func stringsOrEmpty(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}
