package feasibility_test

import (
	"testing"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/feasibility"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/money"
)

func TestConfidenceOrdering(t *testing.T) {
	if feasibility.High.Rank() <= feasibility.Medium.Rank() || feasibility.Medium.Rank() <= feasibility.Low.Rank() {
		t.Fatal("confidence does not order")
	}
	if feasibility.UnknownConfidence.Rank() != 0 || !feasibility.UnknownConfidence.Valid() {
		t.Fatal("unknown confidence mishandled")
	}
	if feasibility.Confidence("certain").Valid() {
		t.Fatal("an invented confidence was accepted")
	}
}

func TestUnknownHelpersNeverClaimKnowledge(t *testing.T) {
	access := feasibility.UnknownAccessAssessment("entry_provider_unavailable")
	if access.Status != feasibility.UnknownAccess || access.EntryMode != feasibility.UnknownEntry || access.Blocks() {
		t.Fatalf("unknown access is not honest: %+v", access)
	}
	budget, err := money.New(100, money.Currency("NGN"))
	if err != nil {
		t.Fatal(err)
	}
	affordability := feasibility.UnknownAffordability(budget, "cost_evidence_unavailable")
	if affordability.Status != feasibility.InsufficientEvidence || affordability.HasEstimate || affordability.Ratio != -1 {
		t.Fatalf("unknown affordability is not honest: %+v", affordability)
	}
	if affordability.Budget.Currency().String() != "NGN" {
		t.Fatal("the traveller's budget currency was lost")
	}
	flight := feasibility.UnknownFlight()
	if flight.Status != feasibility.FlightUnknown || flight.HasPrice {
		t.Fatalf("unknown flight is not honest: %+v", flight)
	}
}

func TestResultValidationRejectsInventedStates(t *testing.T) {
	valid := feasibility.Result{
		Access:        feasibility.Access{Confidence: feasibility.High},
		Affordability: feasibility.Affordability{Confidence: feasibility.Medium},
		Flight:        feasibility.Flight{Confidence: feasibility.Low},
		Overall:       feasibility.Viable,
	}
	if err := valid.Validate(); err != nil {
		t.Fatal(err)
	}
	for name, mutate := range map[string]func(*feasibility.Result){
		"invented overall":    func(r *feasibility.Result) { r.Overall = "amazing" },
		"invented confidence": func(r *feasibility.Result) { r.Access.Confidence = "certain" },
		"empty overall":       func(r *feasibility.Result) { r.Overall = "" },
	} {
		broken := valid
		mutate(&broken)
		if err := broken.Validate(); err == nil {
			t.Fatalf("%s was accepted", name)
		}
	}
}

func TestRankingsOrderStatesSensibly(t *testing.T) {
	easy := feasibility.Access{Status: feasibility.Easy}
	unknown := feasibility.Access{Status: feasibility.UnknownAccess}
	blocked := feasibility.Access{Status: feasibility.NotFeasible}
	if easy.Rank() <= unknown.Rank() || unknown.Rank() <= blocked.Rank() {
		t.Fatal("access ordering is wrong")
	}
	if !blocked.Blocks() || easy.Blocks() {
		t.Fatal("blocking is wrong")
	}
	strong := feasibility.Affordability{Status: feasibility.StrongFit}
	over := feasibility.Affordability{Status: feasibility.OverBudget}
	if strong.Rank() <= over.Rank() || !over.Blocks() || strong.Blocks() {
		t.Fatal("affordability ordering is wrong")
	}
}
