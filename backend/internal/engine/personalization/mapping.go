package personalization

import (
	domain "github.com/SemiuAdesina/nuvro/backend/internal/domain/personalization"
)

var keyBySubject = map[domain.SubjectType]map[string]domain.Key{
	domain.Interest: {
		"coastal": domain.InterestCoastal, "beach": domain.InterestCoastal,
		"food": domain.InterestFood, "culture": domain.InterestCulture,
		"history": domain.InterestCulture, "nature": domain.InterestNature,
		"nightlife": domain.InterestNightlife,
	},
	domain.ActivityKind: {
		"restaurant": domain.InterestFood, "cafe": domain.InterestFood, "market": domain.InterestFood,
		"museum": domain.InterestCulture, "attraction": domain.InterestCulture,
		"park": domain.InterestNature, "viewpoint": domain.InterestNature, "beach": domain.InterestCoastal,
		"nightlife": domain.InterestNightlife,
	},
	domain.StayArea: {
		"quiet": domain.StayQuietArea, "central": domain.StayCentralArea, "lively": domain.StayCentralArea,
	},
	domain.TripPace: {
		"relaxed": domain.PaceRelaxed, "packed": domain.PacePacked,
	},
	domain.TravelBurden: {
		"long_transfers": domain.TransferTolerance, "too_much": domain.TransferTolerance,
	},
}

func KeyFor(subject domain.SubjectType, subjectID string) (domain.Key, bool) {
	byValue, known := keyBySubject[subject]
	if !known {
		return "", false
	}
	key, found := byValue[subjectID]
	return key, found
}

var explicitKeys = map[string]domain.Key{
	"beach": domain.InterestCoastal, "coastal": domain.InterestCoastal,
	"food": domain.InterestFood, "culture": domain.InterestCulture,
	"history": domain.InterestCulture, "nature": domain.InterestNature,
	"nightlife": domain.InterestNightlife,
}

func ExplicitInterestKey(interest string) (domain.Key, bool) {
	key, found := explicitKeys[interest]
	return key, found
}

func ExplicitPaceKey(pace string) (domain.Key, bool) {
	switch pace {
	case "relaxed":
		return domain.PaceRelaxed, true
	case "packed":
		return domain.PacePacked, true
	default:
		return "", false
	}
}

func ExplicitCrowdKey(crowds string) (domain.Key, bool) {
	switch crowds {
	case "quiet":
		return domain.StayQuietArea, true
	case "lively":
		return domain.StayCentralArea, true
	default:
		return "", false
	}
}
