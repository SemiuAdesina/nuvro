package personalization_test

import (
	"testing"
	"time"

	domain "github.com/SemiuAdesina/nuvro/backend/internal/domain/personalization"
	engine "github.com/SemiuAdesina/nuvro/backend/internal/engine/personalization"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
)

var fuzzTypes = []domain.SignalType{
	domain.DestinationSaved, domain.DestinationUnsaved, domain.DestinationSelected,
	domain.ActivityAdded, domain.ActivityRemoved, domain.ActivityCompleted,
	domain.StaySelected, domain.NeighbourhoodSelected, domain.ReplanRequested,
	domain.TripCompleted, domain.FeedbackPositive, domain.FeedbackNegative,
	domain.ExplicitConfirmation,
}

var fuzzSubjects = []domain.SubjectType{
	domain.Destination, domain.ActivityKind, domain.StayArea,
	domain.TripPace, domain.TravelBurden, domain.Interest,
}

var fuzzValues = []string{
	"coastal", "food", "culture", "nature", "nightlife",
	"quiet", "central", "relaxed", "packed", "long_transfers", "unknown", "",
}

func FuzzAggregationInvariants(f *testing.F) {
	f.Add([]byte{0, 5, 0, 1}, uint16(0))
	f.Add([]byte{10, 5, 6, 10, 5, 6, 10, 5, 6}, uint16(400))
	f.Add([]byte{6, 2, 5, 6, 2, 5, 6, 2, 7, 6, 2, 7}, uint16(30))

	learner, err := engine.New(engine.DefaultConfig(), clock.NewFixed(now))
	if err != nil {
		f.Fatal(err)
	}
	f.Fuzz(func(t *testing.T, encoded []byte, spread uint16) {
		signals := decodeSignals(encoded, spread)
		seen := map[domain.Key]bool{}
		for _, preference := range learner.Aggregate(signals) {
			if reason := defectIn(preference); reason != "" {
				t.Fatalf("%q %s: %+v", preference.Key, reason, preference)
			}
			if seen[preference.Key] {
				t.Fatalf("%q was learned twice", preference.Key)
			}
			seen[preference.Key] = true
		}
		for key := range seen {
			for other := range seen {
				if key.Opposes(other) {
					t.Fatalf("%q and %q were both learned", key, other)
				}
			}
		}
	})
}

func defectIn(preference domain.LearnedPreference) string {
	thresholds := engine.DefaultThresholds()
	switch {
	case !preference.Key.Valid() || !preference.Confidence.Valid():
		return "is not a valid preference"
	case preference.EvidenceCount < thresholds.MinimumEvidence:
		return "was learned from too little evidence"
	case preference.Score < thresholds.MediumScore:
		return "was learned below the threshold"
	case preference.Explanation == "" || preference.RulesVersion != domain.RulesVersion:
		return "carries no provenance"
	}
	return ""
}

func decodeSignals(encoded []byte, spread uint16) []domain.Signal {
	signals := make([]domain.Signal, 0, len(encoded)/3)
	for index := 0; index+2 < len(encoded); index += 3 {
		age := time.Duration(int(encoded[index])%int(spread+1)) * 24 * time.Hour
		value := domain.NewSignal("user",
			fuzzTypes[int(encoded[index])%len(fuzzTypes)],
			fuzzSubjects[int(encoded[index+1])%len(fuzzSubjects)],
			fuzzValues[int(encoded[index+2])%len(fuzzValues)],
			now.Add(-age))
		value.SourceEvent = string(rune('a'+index%26)) + value.SubjectID
		signals = append(signals, value)
	}
	return signals
}
