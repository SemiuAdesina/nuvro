package personalization

import "time"

type Insight struct {
	ID            string
	Key           Key
	Category      string
	SummaryCode   string
	Parameters    map[string]string
	Confidence    Confidence
	EvidenceCount int
	Status        Status
	Editable      bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func InsightFrom(preference LearnedPreference) Insight {
	return Insight{
		ID: preference.ID, Key: preference.Key, Category: preference.Key.Category(),
		SummaryCode: "insight." + string(preference.Key),
		Parameters: map[string]string{
			"evidenceCount": itoa(preference.EvidenceCount),
			"explanation":   string(preference.Explanation),
		},
		Confidence: preference.Confidence, EvidenceCount: preference.EvidenceCount,
		Status: preference.Status, Editable: true,
		CreatedAt: preference.FirstObservedAt.UTC(), UpdatedAt: preference.UpdatedAt.UTC(),
	}
}

func itoa(value int) string {
	if value == 0 {
		return "0"
	}
	digits := ""
	for value > 0 {
		digits = string(rune('0'+value%10)) + digits
		value /= 10
	}
	return digits
}
