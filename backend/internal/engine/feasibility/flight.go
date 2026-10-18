package feasibility

import (
	"context"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/destination"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/discovery"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/feasibility"
)

func (e *Engine) flightEvidence(ctx context.Context, intent discovery.Intent, target destination.Destination) (FlightEvidence, bool) {
	if e.flight == nil || intent.Origin.DepartureAirport == "" {
		return FlightEvidence{}, false
	}
	evidence, err := e.flight.Flights(ctx, FlightQuery{
		OriginAirport: intent.Origin.DepartureAirport, DestinationAirport: target.PrimaryAirport,
		Start: intent.Window.Start, End: intent.Window.End, Travelers: intent.Party.Size(),
	})
	if err != nil {
		return FlightEvidence{}, false
	}
	return evidence, true
}

func (e *Engine) costEvidence(ctx context.Context, intent discovery.Intent, target destination.Destination, flight FlightEvidence) (CostEvidence, bool) {
	if e.costs == nil {
		return CostEvidence{}, false
	}
	evidence, err := e.costs.Estimate(ctx, CostQuery{
		DestinationID: target.ID, Nights: intent.Window.Nights(), Travelers: intent.Party.Size(),
		Currency: intent.Budget.Currency().String(), Flight: flight,
	})
	if err != nil {
		return CostEvidence{}, false
	}
	return evidence, true
}

func summarize(evidence FlightEvidence, available bool) feasibility.Flight {
	if !available {
		return feasibility.UnknownFlight()
	}
	if !evidence.Available {
		summary := feasibility.UnknownFlight()
		summary.Status = feasibility.FlightUnavailable
		summary.Confidence = confidenceOrUnknown(evidence.Confidence)
		summary.ObservedAt = evidence.ObservedAt
		return summary
	}
	summary := feasibility.Flight{
		Status: feasibility.FlightAvailable, Stops: evidence.Stops, DurationMinutes: evidence.DurationMinutes,
		Confidence: confidenceOrUnknown(evidence.Confidence), ObservedAt: evidence.ObservedAt,
	}
	if evidence.HasPrice {
		summary.HasPrice = true
		summary.PriceMinor = evidence.Price.Minor()
		summary.PriceCurrency = evidence.Price.Currency().String()
	}
	return summary
}
