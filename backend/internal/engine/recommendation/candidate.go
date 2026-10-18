package recommendation

import (
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/destination"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/feasibility"
)

const RankingVersion = "rank-v1"

type Candidate struct {
	Destination destination.Destination
	Match       destination.MatchScore
	Feasibility feasibility.Result
}

type Recommendation struct {
	Destination destination.Destination
	Match       destination.MatchScore
	Feasibility feasibility.Result
	Reasons     []string
	Summary     string
	Rank        int
}

type Result struct {
	Recommendations []Recommendation
	RankingVersion  string
	Excluded        int
}
