package change_test

import (
	"errors"
	"testing"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/change"
)

func TestThreeStatesAreDistinct(t *testing.T) {
	if !change.Unchanged[string]().IsUnchanged() || change.Unchanged[string]().IsCleared() {
		t.Fatal("unchanged misreported")
	}
	if !change.Clear[string]().IsCleared() || change.Clear[string]().IsUnchanged() {
		t.Fatal("cleared misreported")
	}
	value, ok := change.Set("lisbon").Value()
	if !ok || value != "lisbon" {
		t.Fatal("assigned value lost")
	}
	if _, ok := change.Clear[string]().Value(); ok {
		t.Fatal("cleared reported a value")
	}
}

func TestApplyToRejectsClearingRequiredFields(t *testing.T) {
	current := "EUR"
	if err := change.Unchanged[string]().ApplyTo(&current); err != nil || current != "EUR" {
		t.Fatal("unchanged modified the field")
	}
	if err := change.Set("USD").ApplyTo(&current); err != nil || current != "USD" {
		t.Fatal("assignment failed")
	}
	if err := change.Clear[string]().ApplyTo(&current); !errors.Is(err, change.ErrNotClearable) || current != "USD" {
		t.Fatal("required field was cleared")
	}
}

func TestApplyToOptionalHonoursEveryState(t *testing.T) {
	existing := "LHR"
	target := &existing
	change.Unchanged[string]().ApplyToOptional(&target)
	if target == nil || *target != "LHR" {
		t.Fatal("unchanged modified the optional field")
	}
	change.Set("LOS").ApplyToOptional(&target)
	if target == nil || *target != "LOS" {
		t.Fatal("assignment failed")
	}
	change.Clear[string]().ApplyToOptional(&target)
	if target != nil {
		t.Fatal("clear did not remove the value")
	}
	change.Set("JFK").ApplyToOptional(&target)
	if target == nil || *target != "JFK" || existing != "LHR" {
		t.Fatal("assignment aliased the caller's storage")
	}
}
