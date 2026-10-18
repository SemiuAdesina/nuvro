package retention_test

import (
	"testing"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/retention"
)

func TestEveryRetainedClassStatesHowLongAndWhy(t *testing.T) {
	seen := map[retention.Class]bool{}
	for _, rule := range retention.Policy() {
		if seen[rule.Class] {
			t.Fatalf("%s appears twice in the policy", rule.Class)
		}
		seen[rule.Class] = true
		if rule.MaxAge <= 0 {
			t.Fatalf("%s keeps data forever", rule.Class)
		}
		if rule.Rationale == "" {
			t.Fatalf("%s is kept for %s with no stated reason", rule.Class, rule.MaxAge)
		}
	}
	if len(seen) == 0 {
		t.Fatal("the policy retains nothing at all")
	}
}

func TestPersonalizationSignalsExpireBeforeTheEngineWouldReadThem(t *testing.T) {
	rule, found := retention.RuleFor(retention.PersonalizationSignals)
	if !found {
		t.Fatal("personalization signals are not covered by the policy")
	}
	if rule.MaxAge > 730*24*time.Hour {
		t.Fatalf("signals are kept %s, longer than the two years inference reads", rule.MaxAge)
	}
}

func TestACutoffLooksBackwardsFromNow(t *testing.T) {
	now := time.Date(2026, 9, 9, 12, 0, 0, 0, time.UTC)
	rule := retention.Rule{Class: retention.TripAlerts, MaxAge: 48 * time.Hour}
	if got := rule.CutoffFrom(now); !got.Equal(now.Add(-48 * time.Hour)) {
		t.Fatalf("cutoff = %s, want two days before now", got)
	}
}

func TestAnUnknownClassHasNoRule(t *testing.T) {
	if _, found := retention.RuleFor(retention.Class("invented")); found {
		t.Fatal("an unknown class was given a retention rule")
	}
}
