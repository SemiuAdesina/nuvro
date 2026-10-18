package feasibility

import (
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/money"
)

type AffordabilityStatus string

const (
	StrongFit            AffordabilityStatus = "strong_fit"
	GoodFit              AffordabilityStatus = "good_fit"
	Tight                AffordabilityStatus = "tight"
	OverBudget           AffordabilityStatus = "over_budget"
	InsufficientEvidence AffordabilityStatus = "insufficient_evidence"
)

type CostRange struct {
	Minimum  money.Amount
	Expected money.Amount
	Maximum  money.Amount
}

type Affordability struct {
	Status      AffordabilityStatus
	Budget      money.Amount
	Estimate    CostRange
	HasEstimate bool
	Confidence  Confidence
	Reasons     []string
	ObservedAt  time.Time
	Ratio       int
	Conversion  *money.Conversion
}

func UnknownAffordability(budget money.Amount, reason string) Affordability {
	return Affordability{
		Status: InsufficientEvidence, Budget: budget, Confidence: UnknownConfidence,
		Reasons: []string{reason}, Ratio: -1,
	}
}

func (a Affordability) Rank() int {
	switch a.Status {
	case StrongFit:
		return 4
	case GoodFit:
		return 3
	case Tight:
		return 2
	case InsufficientEvidence:
		return 1
	default:
		return 0
	}
}

func (a Affordability) Blocks() bool { return a.Status == OverBudget }
