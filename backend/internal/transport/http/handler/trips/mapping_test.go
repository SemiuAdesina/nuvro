package trips

import (
	"testing"
	"time"

	"github.com/oapi-codegen/nullable"
	openapi "github.com/oapi-codegen/runtime/types"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/trip"
	"github.com/SemiuAdesina/nuvro/backend/internal/transport/http/api"
	"github.com/SemiuAdesina/nuvro/backend/internal/usecase/plan"
)

func date(day int) openapi.Date {
	return openapi.Date{Time: time.Date(2026, 11, day, 0, 0, 0, 0, time.UTC)}
}

func TestPatchInputMapsThreeStates(t *testing.T) {
	patch, ok := patchInput(api.TripPatch{})
	if !ok || !patch.DestinationID.IsUnchanged() || !patch.Dates.IsUnchanged() || !patch.Adults.IsUnchanged() {
		t.Fatalf("an empty body was not neutral: %+v", patch)
	}
	patch, ok = patchInput(api.TripPatch{DestinationId: nullable.NewNullNullable[string]()})
	if !ok || !patch.DestinationID.IsCleared() {
		t.Fatal("explicit null did not clear the destination")
	}
	patch, ok = patchInput(api.TripPatch{DestinationId: nullable.NewNullableWithValue("azores")})
	if value, present := patch.DestinationID.Value(); !ok || !present || value != "azores" {
		t.Fatal("assignment lost the destination")
	}
}

func TestDatesChangeTogetherOrNotAtAll(t *testing.T) {
	patch, ok := patchInput(api.TripPatch{
		StartDate: nullable.NewNullableWithValue(date(2)), EndDate: nullable.NewNullableWithValue(date(9)),
	})
	value, present := patch.Dates.Value()
	if !ok || !present || value.Start.Day() != 2 || value.End.Day() != 9 {
		t.Fatalf("a full range was not applied: %+v", patch.Dates)
	}
	patch, ok = patchInput(api.TripPatch{
		StartDate: nullable.NewNullNullable[openapi.Date](), EndDate: nullable.NewNullNullable[openapi.Date](),
	})
	if !ok || !patch.Dates.IsCleared() {
		t.Fatal("a paired null did not clear the range")
	}
	for name, body := range map[string]api.TripPatch{
		"start only":    {StartDate: nullable.NewNullableWithValue(date(2))},
		"end only":      {EndDate: nullable.NewNullableWithValue(date(9))},
		"half cleared":  {StartDate: nullable.NewNullNullable[openapi.Date](), EndDate: nullable.NewNullableWithValue(date(9))},
		"half assigned": {StartDate: nullable.NewNullableWithValue(date(2)), EndDate: nullable.NewNullNullable[openapi.Date]()},
	} {
		if _, ok := patchInput(body); ok {
			t.Fatalf("%s was accepted", name)
		}
	}
}

func TestCreationInputIgnoresAbsentOptionalValues(t *testing.T) {
	children := 2
	value, ok := creationInput(api.TripInput{Adults: 3, Children: &children, PreferredCurrency: "EUR"})
	if !ok {
		t.Fatal("a valid body was rejected")
	}
	if value.Adults != 3 || value.Children != 2 || value.Currency != "EUR" {
		t.Fatalf("%+v", value)
	}
	if value.DestinationID != nil || value.OriginAirport != nil || value.StartDate != nil || value.EndDate != nil {
		t.Fatal("absent optional fields became values")
	}
	if value.ID != "" || value.OwnerID != "" || value.Status != "" {
		t.Fatal("the body was allowed to set server-owned fields")
	}
}

func TestResponsesCarryNullsRatherThanOmittedFields(t *testing.T) {
	response := tripResponse(trip.Trip{
		ID: "3f1c9c8e-0e3a-4f0a-9b1f-1a2b3c4d5e6f", OwnerID: "3f1c9c8e-0e3a-4f0a-9b1f-1a2b3c4d5e6f",
		Adults: 1, Currency: "USD", Status: trip.Draft,
	})
	if !response.DestinationId.IsNull() || !response.StartDate.IsNull() || !response.EndDate.IsNull() {
		t.Fatal("absent values were not rendered as null")
	}
	page := pageResponse(plan.Page{})
	if page.Items == nil || len(page.Items) != 0 || page.HasMore || page.NextCursor != nil {
		t.Fatalf("an empty page rendered badly: %+v", page)
	}
	page = pageResponse(plan.Page{HasMore: true, NextCursor: "token"})
	if page.NextCursor == nil || *page.NextCursor != "token" {
		t.Fatal("the cursor was dropped")
	}
}
