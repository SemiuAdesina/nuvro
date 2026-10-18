package feasibility

import "time"

type Confidence string

const (
	High              Confidence = "high"
	Medium            Confidence = "medium"
	Low               Confidence = "low"
	UnknownConfidence Confidence = "unknown"
)

func (c Confidence) Valid() bool {
	switch c {
	case High, Medium, Low, UnknownConfidence:
		return true
	default:
		return false
	}
}

func (c Confidence) Rank() int {
	switch c {
	case High:
		return 3
	case Medium:
		return 2
	case Low:
		return 1
	default:
		return 0
	}
}

type EntryMode string

const (
	VisaFree      EntryMode = "visa_free"
	ETA           EntryMode = "eta"
	EVisa         EntryMode = "evisa"
	VisaOnArrival EntryMode = "visa_on_arrival"
	VisaRequired  EntryMode = "visa_required"
	UnknownEntry  EntryMode = "unknown"
)

type Preparation string

const (
	NoPreparation      Preparation = "none"
	LowPreparation     Preparation = "low"
	MediumPreparation  Preparation = "medium"
	HighPreparation    Preparation = "high"
	UnknownPreparation Preparation = "unknown"
)

type AccessStatus string

const (
	Easy                AccessStatus = "easy"
	Manageable          AccessStatus = "manageable"
	PreparationRequired AccessStatus = "preparation_required"
	NotFeasible         AccessStatus = "not_feasible"
	UnknownAccess       AccessStatus = "unknown"
)

type TransitRisk string

const (
	NoTransitRisk       TransitRisk = "none"
	TransitReviewNeeded TransitRisk = "review_required"
	UnknownTransit      TransitRisk = "unknown"
)

type Access struct {
	Status       AccessStatus
	EntryMode    EntryMode
	Preparation  Preparation
	TransitRisk  TransitRisk
	Confidence   Confidence
	Reasons      []string
	SourceLabel  string
	ObservedAt   time.Time
	PassportUsed string
}

func UnknownAccessAssessment(reason string) Access {
	return Access{
		Status: UnknownAccess, EntryMode: UnknownEntry, Preparation: UnknownPreparation,
		TransitRisk: UnknownTransit, Confidence: UnknownConfidence, Reasons: []string{reason},
	}
}

func (a Access) Rank() int {
	switch a.Status {
	case Easy:
		return 4
	case Manageable:
		return 3
	case PreparationRequired:
		return 2
	case UnknownAccess:
		return 1
	default:
		return 0
	}
}

func (a Access) Blocks() bool { return a.Status == NotFeasible }
