package journey

import (
	uuidlib "github.com/google/uuid"
	openapi "github.com/oapi-codegen/runtime/types"

	budgetdomain "github.com/SemiuAdesina/nuvro/backend/internal/domain/budget"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/itinerary"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/money"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/route"
	"github.com/SemiuAdesina/nuvro/backend/internal/transport/http/api"
)

func ItineraryOf(plan itinerary.Itinerary) api.Itinerary {
	days := make([]api.ItineraryDay, 0, len(plan.Days()))
	for _, day := range plan.Days() {
		items := make([]api.ItineraryItem, 0, len(day.Items))
		for _, item := range day.Items {
			items = append(items, ItemOf(item))
		}
		days = append(days, api.ItineraryDay{
			Day: day.Number, Date: openapi.Date{Time: day.Date.UTC()}, Items: items,
		})
	}
	value := api.Itinerary{
		TripId: uuid(plan.TripID), Timezone: plan.Timezone, Version: plan.Version,
		AlgorithmVersion: plan.AlgorithmVersion, GeneratedAt: plan.GeneratedAt.UTC(), Days: days,
	}
	if plan.ID != "" {
		identifier := uuid(plan.ID)
		value.Id = &identifier
	}
	return value
}

func ItemOf(item itinerary.Item) api.ItineraryItem {
	value := api.ItineraryItem{
		Id: uuid(item.ID), Day: item.Day, StartAt: item.StartAt.UTC(), EndAt: item.EndAt.UTC(),
		Kind: api.ItineraryItemKind(item.Kind), Title: item.Title,
		WeatherImpact: api.ItineraryItemWeatherImpact(item.WeatherImpact),
		Status:        api.ItineraryItemStatus(item.Status), Source: api.ItineraryItemSource(item.Source),
		UserLocked: item.UserLocked, ReservationLocked: item.ReservationLocked,
	}
	if item.PlaceID != "" {
		placeID := item.PlaceID
		value.PlaceId = &placeID
	}
	if item.Category != "" {
		category := string(item.Category)
		value.Category = &category
	}
	if !item.Location.IsZero() {
		value.Location = &api.GeoPoint{
			Latitude: float32(item.Location.Latitude), Longitude: float32(item.Location.Longitude),
		}
	}
	if item.RouteFromPrevious != nil {
		value.RouteFromPrevious = routeOf(*item.RouteFromPrevious)
	}
	if item.EstimatedCostMinor > 0 && item.CostCurrency != "" {
		value.EstimatedCost = &api.Money{AmountMinor: item.EstimatedCostMinor, Currency: item.CostCurrency}
	}
	return value
}

func routeOf(evidence route.Evidence) *api.RouteSummary {
	summary := api.RouteSummary{
		Status: api.RouteSummaryStatus(evidence.Status), Mode: api.RouteSummaryMode(evidence.Mode),
	}
	if evidence.Confidence != "" {
		confidence := api.RouteSummaryConfidence(evidence.Confidence)
		summary.Confidence = &confidence
	}
	if evidence.Known() {
		distance := evidence.DistanceMeters
		duration := evidence.DurationSeconds
		summary.DistanceMeters = &distance
		summary.DurationSeconds = &duration
	}
	return &summary
}

func ViolationsOf(values []itinerary.Violation) []api.Violation {
	violations := make([]api.Violation, 0, len(values))
	for _, value := range values {
		violation := api.Violation{Code: api.ViolationCode(value.Code)}
		if value.ItemID != "" {
			itemID := value.ItemID
			violation.ItemId = &itemID
		}
		if value.Detail != "" {
			detail := value.Detail
			violation.Detail = &detail
		}
		violations = append(violations, violation)
	}
	return violations
}

func budgetOf(result budgetdomain.Trip) api.TripBudget {
	lines := make([]api.BudgetLine, 0, len(result.Lines))
	for _, line := range result.Lines {
		lines = append(lines, lineOf(line))
	}
	value := api.TripBudget{
		Currency: result.Currency.String(), Categories: lines,
		Confidence: api.TripBudgetConfidence(result.Confidence), CalculatedAt: result.CalculatedAt.UTC(),
	}
	if result.HasBudget {
		value.TripBudget = amountOf(result.Budget)
	}
	if result.HasEstimate {
		value.EstimatedTotal = rangeOf(result.Estimated)
		value.BookedTotal = amountOf(result.Booked)
	}
	if result.HasRemaining {
		value.Remaining = amountOf(result.Remaining)
	}
	return value
}

func lineOf(line budgetdomain.Line) api.BudgetLine {
	value := api.BudgetLine{
		Category: api.BudgetLineCategory(line.Category), Confidence: api.BudgetLineConfidence(line.Confidence),
	}
	if line.HasEstimate {
		value.Estimated = rangeOf(line.Estimated)
	}
	if line.HasBooked {
		value.Booked = amountOf(line.Booked)
	}
	if line.Source != "" {
		source := line.Source
		value.Source = &source
	}
	if !line.ObservedAt.IsZero() {
		observedAt := line.ObservedAt.UTC()
		value.ObservedAt = &observedAt
	}
	return value
}

func rangeOf(amounts budgetdomain.Range) *api.CostRange {
	return &api.CostRange{
		Minimum:  *amountOf(amounts.Minimum),
		Expected: *amountOf(amounts.Expected),
		Maximum:  *amountOf(amounts.Maximum),
	}
}

func amountOf(amount money.Amount) *api.Money {
	return &api.Money{AmountMinor: amount.Minor(), Currency: amount.Currency().String()}
}

func uuid(value string) openapi.UUID {
	parsed, err := uuidlib.Parse(value)
	if err != nil {
		return openapi.UUID{}
	}
	return parsed
}

func DayOf(day itinerary.Day) api.ItineraryDay {
	items := make([]api.ItineraryItem, 0, len(day.Items))
	for _, item := range day.Items {
		items = append(items, ItemOf(item))
	}
	return api.ItineraryDay{Day: day.Number, Date: openapi.Date{Time: day.Date.UTC()}, Items: items}
}
