package personalization

import (
	"slices"
	"time"

	domain "github.com/SemiuAdesina/nuvro/backend/internal/domain/personalization"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
)

type Engine struct {
	config Config
	timer  clock.Clock
}

func New(config Config, timer clock.Clock) (*Engine, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	return &Engine{config: config, timer: timer}, nil
}

func (e *Engine) Version() string { return e.config.Version }

type Evidence struct {
	Key      domain.Key
	Positive int
	Negative int
	Count    int
	First    time.Time
	Last     time.Time
	Source   domain.ExplanationCode
}

func (e *Engine) Aggregate(signals []domain.Signal) []domain.LearnedPreference {
	now := e.timer.Now().UTC()
	byKey := map[domain.Key]*Evidence{}
	for _, signal := range signals {
		if signal.Validate() != nil {
			continue
		}
		key, mapped := KeyFor(signal.Subject, signal.SubjectID)
		if !mapped {
			continue
		}
		e.accumulate(byKey, key, signal, now)
	}
	return e.preferences(byKey, now)
}

func (e *Engine) accumulate(byKey map[domain.Key]*Evidence, key domain.Key, signal domain.Signal, now time.Time) {
	entry, exists := byKey[key]
	if !exists {
		entry = &Evidence{Key: key, First: signal.OccurredAt, Last: signal.OccurredAt}
		byKey[key] = entry
	}
	weight := signal.Strength.Weight() * e.config.Decay.Factor(ClassOf(key), now.Sub(signal.OccurredAt)) / 100
	if signal.Type.Negative() {
		entry.Negative += weight
	} else {
		entry.Positive += weight
	}
	entry.Count++
	if signal.OccurredAt.Before(entry.First) {
		entry.First = signal.OccurredAt
	}
	if signal.OccurredAt.After(entry.Last) {
		entry.Last = signal.OccurredAt
	}
	if explanation := explanationFor(signal.Type); rankExplanation(explanation) > rankExplanation(entry.Source) {
		entry.Source = explanation
	}
}

func (e *Engine) preferences(byKey map[domain.Key]*Evidence, now time.Time) []domain.LearnedPreference {
	learned := make([]domain.LearnedPreference, 0, len(byKey))
	for _, entry := range byKey {
		score := entry.Positive - entry.Negative
		if entry.Count < e.config.Thresholds.MinimumEvidence || score < e.config.Thresholds.MediumScore {
			continue
		}
		learned = append(learned, domain.LearnedPreference{
			UserID: "", Key: entry.Key, Confidence: e.confidence(score), Score: score,
			EvidenceCount: entry.Count, Explanation: entry.Source, Status: domain.Active,
			RulesVersion: e.config.Version, FirstObservedAt: entry.First,
			LastObservedAt: entry.Last, UpdatedAt: now,
		})
	}
	slices.SortStableFunc(learned, func(left, right domain.LearnedPreference) int {
		if order := right.Score - left.Score; order != 0 {
			return order
		}
		return slices.Compare([]string{string(left.Key)}, []string{string(right.Key)})
	})
	return e.resolveConflicts(learned)
}

func (e *Engine) resolveConflicts(learned []domain.LearnedPreference) []domain.LearnedPreference {
	kept := make([]domain.LearnedPreference, 0, len(learned))
	for _, candidate := range learned {
		conflicting := false
		for _, existing := range kept {
			if existing.Key.Opposes(candidate.Key) {
				conflicting = true
				break
			}
		}
		if !conflicting {
			kept = append(kept, candidate)
		}
	}
	return kept
}

func (e *Engine) confidence(score int) domain.Confidence {
	switch {
	case score >= e.config.Thresholds.HighScore:
		return domain.High
	case score >= e.config.Thresholds.MediumScore:
		return domain.MediumConfidence
	default:
		return domain.Low
	}
}

func explanationFor(signalType domain.SignalType) domain.ExplanationCode {
	switch signalType {
	case domain.ExplicitConfirmation:
		return domain.UserConfirmation
	case domain.FeedbackPositive, domain.FeedbackNegative:
		return domain.PostTripFeedback
	case domain.ReplanRequested:
		return domain.RepeatedReplanning
	case domain.DestinationSaved, domain.DestinationSelected:
		return domain.RepeatedSaves
	default:
		return domain.RepeatedChoices
	}
}

func rankExplanation(code domain.ExplanationCode) int {
	switch code {
	case domain.UserConfirmation:
		return 4
	case domain.PostTripFeedback:
		return 3
	case domain.RepeatedSaves, domain.RepeatedReplanning:
		return 2
	case domain.RepeatedChoices:
		return 1
	default:
		return 0
	}
}

func (e *Engine) Context(userID string, preferences user.Preferences, learned []domain.LearnedPreference, enabled bool) domain.Context {
	context := domain.EmptyContext(userID, enabled, e.timer.Now())
	context.RulesVersion = e.config.Version
	context.Explicit = explicitPreferences(preferences)
	for _, preference := range learned {
		if preference.Status == domain.Dismissed || preference.Status == domain.Suppressed {
			context.Suppressed = append(context.Suppressed, preference.Key)
			continue
		}
		if !enabled || !preference.Influences() {
			continue
		}
		context.Learned = append(context.Learned, domain.Preference{
			Key: preference.Key, Origin: originOf(preference.Status),
			Confidence: preference.Confidence,
			Weight:     domain.WeightFor(originOf(preference.Status), preference.Confidence),
		})
	}
	return context
}

func originOf(status domain.Status) domain.Origin {
	if status == domain.Confirmed {
		return domain.ConfirmedLearned
	}
	return domain.InferredLearned
}

func explicitPreferences(preferences user.Preferences) []domain.Preference {
	explicit := make([]domain.Preference, 0)
	seen := map[domain.Key]struct{}{}
	add := func(key domain.Key, found bool) {
		if !found {
			return
		}
		if _, duplicate := seen[key]; duplicate {
			return
		}
		seen[key] = struct{}{}
		explicit = append(explicit, domain.Preference{
			Key: key, Origin: domain.ExplicitProfile, Confidence: domain.High,
			Weight: domain.ExplicitWeight,
		})
	}
	for _, interest := range preferences.Interests {
		add(ExplicitInterestKey(interest))
	}
	add(ExplicitPaceKey(preferences.Pace))
	add(ExplicitCrowdKey(preferences.Crowds))
	slices.SortStableFunc(explicit, func(left, right domain.Preference) int {
		return slices.Compare([]string{string(left.Key)}, []string{string(right.Key)})
	})
	return explicit
}
