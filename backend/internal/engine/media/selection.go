package media

import (
	"context"
	"slices"
	"strings"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/destination"
)

const MinimumWidth = 640
const MinimumHeight = 480
const SelectionTTL = 6 * time.Hour

type Selector struct{ reader CandidateReader }

func NewSelector(reader CandidateReader) *Selector { return &Selector{reader: reader} }

func (s *Selector) Configured() bool { return s.reader != nil }

func (s *Selector) Select(ctx context.Context, context Context, limit int) []destination.Media {
	if s.reader == nil || limit < 1 {
		return nil
	}
	candidates, err := s.reader.Candidates(ctx, context)
	if err != nil {
		return nil
	}
	return Choose(candidates, context, limit)
}

func Choose(candidates []Candidate, context Context, limit int) []destination.Media {
	accepted := make([]Candidate, 0, len(candidates))
	seen := make(map[string]struct{}, len(candidates))
	for _, candidate := range candidates {
		if !acceptable(candidate, context) {
			continue
		}
		fingerprint := strings.ToLower(candidate.Media.URL)
		if _, duplicate := seen[fingerprint]; duplicate {
			continue
		}
		seen[fingerprint] = struct{}{}
		accepted = append(accepted, candidate)
	}
	slices.SortStableFunc(accepted, func(left, right Candidate) int {
		if order := score(right, context) - score(left, context); order != 0 {
			return order
		}
		return strings.Compare(left.ProviderReference, right.ProviderReference)
	})
	if len(accepted) > limit {
		accepted = accepted[:limit]
	}
	selected := make([]destination.Media, 0, len(accepted))
	for _, candidate := range accepted {
		selected = append(selected, candidate.Media)
	}
	return selected
}

func acceptable(candidate Candidate, context Context) bool {
	media := candidate.Media
	if media.URL == "" || len(media.Attribution) == 0 {
		return false
	}
	if media.Width < MinimumWidth || media.Height < MinimumHeight {
		return false
	}
	if media.ExpiresAt != nil && !media.ExpiresAt.After(time.Now().UTC()) {
		return false
	}
	if context.Landscape && media.Width <= media.Height {
		return false
	}
	return true
}

func score(candidate Candidate, context Context) int {
	media := candidate.Media
	total := 0
	if context.Landscape {
		total += aspectScore(media.AspectRatio, 1.5)
	} else {
		total += aspectScore(media.AspectRatio, 1.0)
	}
	total += min(media.Width/320, 8)
	if media.AltText != "" {
		total += 3
	}
	if candidate.Season != "" && candidate.Season == seasonOf(context.StartDate) {
		total += 6
	}
	return total
}

func aspectScore(actual, wanted float64) int {
	if actual <= 0 {
		return 0
	}
	difference := actual - wanted
	if difference < 0 {
		difference = -difference
	}
	return max(0, 10-int(difference*10))
}

func seasonOf(date time.Time) string {
	if date.IsZero() {
		return ""
	}
	switch date.UTC().Month() {
	case time.December, time.January, time.February:
		return "winter"
	case time.March, time.April, time.May:
		return "spring"
	case time.June, time.July, time.August:
		return "summer"
	default:
		return "autumn"
	}
}
