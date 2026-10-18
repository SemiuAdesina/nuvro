package recommendation

import (
	"slices"
	"strings"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/feasibility"
)

func rank(candidates []Candidate) []Candidate {
	ordered := slices.Clone(candidates)
	slices.SortStableFunc(ordered, compare)
	return ordered
}

func compare(left, right Candidate) int {
	if order := right.Match.Score - left.Match.Score; order != 0 {
		return order
	}
	if order := right.Feasibility.Rank() - left.Feasibility.Rank(); order != 0 {
		return order
	}
	if order := right.Feasibility.Access.Rank() - left.Feasibility.Access.Rank(); order != 0 {
		return order
	}
	if order := costOrder(left, right); order != 0 {
		return order
	}
	return strings.Compare(left.Destination.ID, right.Destination.ID)
}

func costOrder(left, right Candidate) int {
	leftRatio, leftKnown := costRatio(left)
	rightRatio, rightKnown := costRatio(right)
	switch {
	case leftKnown && !rightKnown:
		return -1
	case !leftKnown && rightKnown:
		return 1
	case !leftKnown && !rightKnown:
		return 0
	default:
		return leftRatio - rightRatio
	}
}

func costRatio(candidate Candidate) (int, bool) {
	affordability := candidate.Feasibility.Affordability
	if !affordability.HasEstimate || affordability.Ratio < 0 {
		return 0, false
	}
	return affordability.Ratio, true
}

func excluded(candidate Candidate) bool {
	return candidate.Feasibility.Overall == feasibility.Infeasible
}
