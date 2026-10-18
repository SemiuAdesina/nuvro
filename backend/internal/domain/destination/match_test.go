package destination_test

import (
	"testing"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/destination"
)

func TestMatchScoreInvariants(t *testing.T) {
	valid := destination.MatchScore{Score: 75, WeightsVersion: "v1", EvidenceVersion: "fixture"}
	if err := valid.Validate(); err != nil {
		t.Fatal(err)
	}
	for _, mutate := range []func(*destination.MatchScore){
		func(m *destination.MatchScore) { m.Score = 101 },
		func(m *destination.MatchScore) { m.Score = -1 },
		func(m *destination.MatchScore) { m.WeightsVersion = "" },
		func(m *destination.MatchScore) {
			m.Components = []destination.MatchComponent{{AwardedPoints: 11, PossiblePoints: 10, Assessment: destination.Good}}
		},
		func(m *destination.MatchScore) { m.Components = []destination.MatchComponent{{Assessment: "invented"}} },
	} {
		value := valid
		mutate(&value)
		if value.Validate() == nil {
			t.Fatalf("accepted %+v", value)
		}
	}
}
