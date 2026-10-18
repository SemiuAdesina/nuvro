package discover

import (
	"context"
	"slices"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/discovery"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/personalization"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
)

var interestByKey = map[personalization.Key]string{
	personalization.InterestCoastal:   "beach",
	personalization.InterestFood:      "food",
	personalization.InterestCulture:   "culture",
	personalization.InterestNature:    "nature",
	personalization.InterestNightlife: "nightlife",
}

func (s *Service) personalize(ctx context.Context, principal user.Principal, intent discovery.Intent) discovery.Intent {
	if s.personalization == nil {
		return intent
	}
	learned, err := s.personalization.Context(ctx, principal)
	if err != nil || !learned.HistoryLearningEnabled {
		return intent
	}
	intent.Preferences.Interests = applyInterests(intent.Preferences.Interests, learned)
	intent.Preferences.Pace = applyPace(intent.Preferences.Pace, learned)
	intent.Preferences.Crowds = applyCrowds(intent.Preferences.Crowds, learned)
	return intent
}

func applyInterests(interests []string, learned personalization.Context) []string {
	extended := slices.Clone(interests)
	for key, interest := range interestByKey {
		if slices.Contains(extended, interest) {
			continue
		}
		if origin, found := learned.OriginOf(key); found && origin != personalization.ExplicitProfile {
			extended = append(extended, interest)
		}
	}
	return discovery.NormalizeTags(extended)
}

func applyPace(pace string, learned personalization.Context) string {
	if pace != "" && pace != "any" {
		return pace
	}
	if learned.Holds(personalization.PaceRelaxed) {
		return "relaxed"
	}
	if learned.Holds(personalization.PacePacked) {
		return "packed"
	}
	return pace
}

func applyCrowds(crowds string, learned personalization.Context) string {
	if crowds != "" && crowds != "any" {
		return crowds
	}
	if learned.Holds(personalization.StayQuietArea) {
		return "quiet"
	}
	if learned.Holds(personalization.StayCentralArea) {
		return "lively"
	}
	return crowds
}
