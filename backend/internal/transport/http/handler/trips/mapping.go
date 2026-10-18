package trips

import (
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/oapi-codegen/nullable"
	openapi "github.com/oapi-codegen/runtime/types"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/change"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/trip"
	"github.com/SemiuAdesina/nuvro/backend/internal/transport/http/api"
	"github.com/SemiuAdesina/nuvro/backend/internal/usecase/plan"
)

func tripResponse(value trip.Trip) api.Trip {
	identifier, _ := uuid.Parse(value.ID)
	owner, _ := uuid.Parse(value.OwnerID)
	return api.Trip{
		Id: identifier, OwnerId: owner, Status: api.TripStatus(value.Status),
		DestinationId: nullableText(value.DestinationID), OriginAirport: nullableText(value.OriginAirport),
		StartDate: nullableDate(value.StartDate), EndDate: nullableDate(value.EndDate),
		Adults: int(value.Adults), Children: int(value.Children), PreferredCurrency: value.Currency,
		CreatedAt: value.CreatedAt.UTC(), UpdatedAt: value.UpdatedAt.UTC(),
	}
}

func pageResponse(page plan.Page) api.TripPage {
	items := make([]api.Trip, 0, len(page.Trips))
	for _, value := range page.Trips {
		items = append(items, tripResponse(value))
	}
	response := api.TripPage{Items: items, HasMore: page.HasMore}
	if page.NextCursor != "" {
		cursor := page.NextCursor
		response.NextCursor = &cursor
	}
	return response
}

func creationInput(body api.TripInput) (trip.Trip, bool) {
	adults, ok := bounded(body.Adults)
	if !ok {
		return trip.Trip{}, false
	}
	value := trip.Trip{Adults: adults, Currency: body.PreferredCurrency}
	if body.Children != nil {
		children, ok := bounded(*body.Children)
		if !ok {
			return trip.Trip{}, false
		}
		value.Children = children
	}
	value.DestinationID = optionalText(body.DestinationId)
	value.OriginAirport = optionalText(body.OriginAirport)
	value.StartDate = optionalDate(body.StartDate)
	value.EndDate = optionalDate(body.EndDate)
	return value, true
}

func bounded(value int) (int32, bool) {
	if value < math.MinInt32 || value > math.MaxInt32 {
		return 0, false
	}
	return int32(value), true
}

func patchInput(body api.TripPatch) (trip.Patch, bool) {
	adults, adultsOK := countChange(body.Adults)
	children, childrenOK := countChange(body.Children)
	dates, datesOK := dateChange(body.StartDate, body.EndDate)
	if !adultsOK || !childrenOK || !datesOK {
		return trip.Patch{}, false
	}
	return trip.Patch{
		DestinationID: textChange(body.DestinationId),
		OriginAirport: textChange(body.OriginAirport),
		Adults:        adults,
		Children:      children,
		Currency:      valueChange(body.PreferredCurrency),
		Dates:         dates,
	}, true
}

func dateChange(start, end nullable.Nullable[openapi.Date]) (change.Change[trip.Dates], bool) {
	if !start.IsSpecified() && !end.IsSpecified() {
		return change.Unchanged[trip.Dates](), true
	}
	if !start.IsSpecified() || !end.IsSpecified() || start.IsNull() != end.IsNull() {
		return change.Unchanged[trip.Dates](), false
	}
	if start.IsNull() {
		return change.Clear[trip.Dates](), true
	}
	from, err := start.Get()
	if err != nil {
		return change.Unchanged[trip.Dates](), false
	}
	to, err := end.Get()
	if err != nil {
		return change.Unchanged[trip.Dates](), false
	}
	return change.Set(trip.Dates{Start: from.UTC(), End: to.UTC()}), true
}

func textChange(value nullable.Nullable[string]) change.Change[string] {
	if !value.IsSpecified() {
		return change.Unchanged[string]()
	}
	if value.IsNull() {
		return change.Clear[string]()
	}
	assigned, err := value.Get()
	if err != nil {
		return change.Unchanged[string]()
	}
	return change.Set(assigned)
}

func valueChange(value *string) change.Change[string] {
	if value == nil {
		return change.Unchanged[string]()
	}
	return change.Set(*value)
}

func countChange(value *int) (change.Change[int32], bool) {
	if value == nil {
		return change.Unchanged[int32](), true
	}
	count, ok := bounded(*value)
	if !ok {
		return change.Unchanged[int32](), false
	}
	return change.Set(count), true
}

func optionalText(value nullable.Nullable[string]) *string {
	assigned, err := value.Get()
	if err != nil || !value.IsSpecified() || value.IsNull() {
		return nil
	}
	return &assigned
}

func optionalDate(value nullable.Nullable[openapi.Date]) *time.Time {
	assigned, err := value.Get()
	if err != nil || !value.IsSpecified() || value.IsNull() {
		return nil
	}
	instant := assigned.UTC()
	return &instant
}

func nullableText(value *string) nullable.Nullable[string] {
	if value == nil {
		return nullable.NewNullNullable[string]()
	}
	return nullable.NewNullableWithValue(*value)
}

func nullableDate(value *time.Time) nullable.Nullable[openapi.Date] {
	if value == nil {
		return nullable.NewNullNullable[openapi.Date]()
	}
	return nullable.NewNullableWithValue(openapi.Date{Time: value.UTC()})
}
