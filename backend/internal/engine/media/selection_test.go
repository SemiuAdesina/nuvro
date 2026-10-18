package media_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/destination"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/evidence"
	"github.com/SemiuAdesina/nuvro/backend/internal/engine/media"
)

var summer = time.Date(2026, 7, 14, 0, 0, 0, 0, time.UTC)

type reader struct {
	candidates []media.Candidate
	err        error
}

func (r reader) Candidates(context.Context, media.Context) ([]media.Candidate, error) {
	return r.candidates, r.err
}

func asset(url string, width, height int, credited bool) destination.Media {
	value := destination.Media{
		URL: url, Width: width, Height: height, AltText: "A view",
	}
	if height > 0 {
		value.AspectRatio = float64(width) / float64(height)
	}
	if credited {
		value.Attribution = []evidence.Attribution{{Text: "A Photographer", Provider: "google_places"}}
	}
	return value
}

func candidate(url string, width, height int, credited bool, season string) media.Candidate {
	return media.Candidate{Media: asset(url, width, height, credited), ProviderReference: url, Season: season}
}

func TestUncreditedOrUndersizedPhotosAreRejected(t *testing.T) {
	selected := media.Choose([]media.Candidate{
		candidate("good", 1600, 900, true, ""),
		candidate("uncredited", 1600, 900, false, ""),
		candidate("tiny", 320, 240, true, ""),
		candidate("no-url", 0, 0, true, ""),
	}, media.Context{Landscape: true}, 5)
	if len(selected) != 1 || selected[0].URL != "good" {
		t.Fatalf("selected %v", urls(selected))
	}
}

func TestLandscapeContextRejectsPortraitPhotos(t *testing.T) {
	selected := media.Choose([]media.Candidate{
		candidate("portrait", 900, 1600, true, ""),
		candidate("landscape", 1600, 900, true, ""),
	}, media.Context{Landscape: true}, 5)
	if len(selected) != 1 || selected[0].URL != "landscape" {
		t.Fatalf("selected %v", urls(selected))
	}
	both := media.Choose([]media.Candidate{
		candidate("portrait", 900, 1600, true, ""),
		candidate("landscape", 1600, 900, true, ""),
	}, media.Context{}, 5)
	if len(both) != 2 {
		t.Fatal("a square-ish context should keep both orientations")
	}
}

func TestExpiredPhotosAreNeverOffered(t *testing.T) {
	expired := candidate("expired", 1600, 900, true, "")
	past := time.Now().UTC().Add(-time.Hour)
	expired.Media.ExpiresAt = &past
	valid := candidate("valid", 1600, 900, true, "")
	future := time.Now().UTC().Add(time.Hour)
	valid.Media.ExpiresAt = &future
	selected := media.Choose([]media.Candidate{expired, valid}, media.Context{Landscape: true}, 5)
	if len(selected) != 1 || selected[0].URL != "valid" {
		t.Fatalf("selected %v", urls(selected))
	}
}

func TestDuplicatesAreRemoved(t *testing.T) {
	selected := media.Choose([]media.Candidate{
		candidate("same", 1600, 900, true, ""),
		candidate("SAME", 1600, 900, true, ""),
	}, media.Context{Landscape: true}, 5)
	if len(selected) != 1 {
		t.Fatalf("duplicates survived: %v", urls(selected))
	}
}

func TestAttributionSurvivesSelection(t *testing.T) {
	selected := media.Choose([]media.Candidate{candidate("a", 1600, 900, true, "")}, media.Context{Landscape: true}, 5)
	if len(selected[0].Attribution) != 1 || selected[0].Attribution[0].Provider == "" {
		t.Fatalf("attribution lost: %+v", selected[0].Attribution)
	}
}

func TestSeasonalRelevanceIsUsedOnlyWhenProvidedBySource(t *testing.T) {
	selected := media.Choose([]media.Candidate{
		candidate("winter", 1600, 900, true, "winter"),
		candidate("summer", 1600, 900, true, "summer"),
	}, media.Context{Landscape: true, StartDate: summer}, 5)
	if selected[0].URL != "summer" {
		t.Fatalf("seasonal relevance ignored: %v", urls(selected))
	}
	unlabelled := media.Choose([]media.Candidate{
		candidate("first", 1600, 900, true, ""),
		candidate("second", 1600, 900, true, ""),
	}, media.Context{Landscape: true, StartDate: summer}, 5)
	if len(unlabelled) != 2 {
		t.Fatal("unlabelled candidates were dropped")
	}
}

func TestSelectionIsStableForTheSameInput(t *testing.T) {
	candidates := []media.Candidate{
		candidate("a", 1600, 900, true, ""),
		candidate("b", 1920, 1080, true, ""),
		candidate("c", 2400, 1350, true, ""),
	}
	first := urls(media.Choose(candidates, media.Context{Landscape: true}, 2))
	for range 20 {
		again := urls(media.Choose(candidates, media.Context{Landscape: true}, 2))
		for index := range again {
			if again[index] != first[index] {
				t.Fatalf("selection changed between identical calls: %v then %v", first, again)
			}
		}
	}
}

func TestSelectorHandlesNoProviderAndProviderFailure(t *testing.T) {
	unconfigured := media.NewSelector(nil)
	if unconfigured.Configured() {
		t.Fatal("an unconfigured selector claimed to be configured")
	}
	if selected := unconfigured.Select(t.Context(), media.Context{}, 5); selected != nil {
		t.Fatal("an unconfigured selector produced media")
	}
	failing := media.NewSelector(reader{err: errors.New("provider down")})
	if selected := failing.Select(t.Context(), media.Context{}, 5); len(selected) != 0 {
		t.Fatal("a failing provider produced media")
	}
	working := media.NewSelector(reader{candidates: []media.Candidate{candidate("a", 1600, 900, true, "")}})
	if len(working.Select(t.Context(), media.Context{Landscape: true}, 5)) != 1 {
		t.Fatal("a working provider produced nothing")
	}
	if working.Select(t.Context(), media.Context{Landscape: true}, 0) != nil {
		t.Fatal("a zero limit produced media")
	}
}

func urls(assets []destination.Media) []string {
	values := make([]string, 0, len(assets))
	for _, asset := range assets {
		values = append(values, asset.URL)
	}
	return values
}
