package place_test

import (
	"strings"
	"testing"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/geo"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/place"
)

var monday = time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC)

func valid() place.Place {
	return place.Place{
		ID: "fixture:a", Name: "Museu", Category: place.Museum,
		Location: geo.Point{Latitude: 32.6478, Longitude: -16.9048}, Status: place.Open,
		Rating: 4.5, ReviewCount: 100, PriceLevel: 2,
	}
}

func TestPlaceValidation(t *testing.T) {
	if err := valid().Validate(); err != nil {
		t.Fatal(err)
	}
	for name, mutate := range map[string]func(*place.Place){
		"no id":           func(p *place.Place) { p.ID = "" },
		"no name":         func(p *place.Place) { p.Name = "" },
		"invented type":   func(p *place.Place) { p.Category = "spaceport" },
		"bad coordinates": func(p *place.Place) { p.Location = geo.Point{Latitude: 200} },
		"absurd rating":   func(p *place.Place) { p.Rating = 9 },
		"negative count":  func(p *place.Place) { p.ReviewCount = -1 },
		"absurd price":    func(p *place.Place) { p.PriceLevel = 9 },
		"invented status": func(p *place.Place) { p.Status = "maybe" },
	} {
		record := valid()
		mutate(&record)
		if record.Validate() == nil {
			t.Fatalf("%s was accepted", name)
		}
	}
}

func TestCategoriesClassifyExposure(t *testing.T) {
	if !place.Beach.Outdoor() || place.Beach.Indoor() {
		t.Fatal("a beach is not outdoors")
	}
	if !place.Museum.Indoor() || place.Museum.Outdoor() {
		t.Fatal("a museum is not indoors")
	}
	if place.ParseCategory("spaceport") != place.Other {
		t.Fatal("an unknown provider type did not fall back to other")
	}
	if place.ParseCategory("museum") != place.Museum {
		t.Fatal("a known category did not parse")
	}
}

func TestDurationsHaveSensibleDefaults(t *testing.T) {
	explicit := valid()
	explicit.TypicalMinutes = 45
	if explicit.Duration() != 45*time.Minute {
		t.Fatalf("explicit duration ignored: %v", explicit.Duration())
	}
	if valid().Duration() <= 0 {
		t.Fatal("a place with no stated duration has no default")
	}
	beach := valid()
	beach.Category = place.Beach
	cafe := valid()
	cafe.Category = place.Cafe
	if beach.Duration() <= cafe.Duration() {
		t.Fatal("a beach visit should take longer than a coffee")
	}
}

func TestOpeningHoursAnswerOnlyWhatTheyKnow(t *testing.T) {
	hours := place.OpeningHours{
		Known:     true,
		Intervals: []place.Interval{{Weekday: time.Monday, Opens: 9 * 60, Closes: 17 * 60}},
	}
	if hours.OpenDuring(monday.Add(10*time.Hour), monday.Add(12*time.Hour)) != place.Open {
		t.Fatal("a visit inside opening hours was not open")
	}
	if hours.OpenDuring(monday.Add(18*time.Hour), monday.Add(19*time.Hour)) != place.Closed {
		t.Fatal("a visit after closing was not closed")
	}
	if hours.OpenDuring(monday.Add(16*time.Hour), monday.Add(18*time.Hour)) != place.Closed {
		t.Fatal("a visit overrunning closing time was accepted")
	}
	tuesday := monday.AddDate(0, 0, 1)
	if hours.OpenDuring(tuesday.Add(10*time.Hour), tuesday.Add(12*time.Hour)) != place.Closed {
		t.Fatal("a day with no interval was treated as open")
	}
	unknown := place.OpeningHours{}
	if unknown.OpenDuring(monday.Add(10*time.Hour), monday.Add(11*time.Hour)) != place.UnknownStatus {
		t.Fatal("unknown hours claimed knowledge")
	}
	if hours.OpenDuring(monday.Add(23*time.Hour), monday.AddDate(0, 0, 1).Add(time.Hour)) != place.UnknownStatus {
		t.Fatal("a visit spanning midnight was answered rather than deferred")
	}
}

func TestReferenceLookup(t *testing.T) {
	record := valid()
	record.References = []place.Reference{{Provider: "google_places", ExternalID: "X"}}
	if reference, found := record.Reference("google_places"); !found || reference.ExternalID != "X" {
		t.Fatal("provider identity lost")
	}
	if _, found := record.Reference("someone_else"); found {
		t.Fatal("a reference from another provider was returned")
	}
}

func FuzzParseCategory(f *testing.F) {
	f.Add("museum")
	f.Add("")
	f.Add(strings.Repeat("x", 200))
	f.Fuzz(func(t *testing.T, raw string) {
		category := place.ParseCategory(raw)
		if !category.Valid() {
			t.Fatalf("%q produced an invalid category %q", raw, category)
		}
		if category != place.Other && string(category) != raw {
			t.Fatalf("%q became %q", raw, category)
		}
	})
}

func FuzzOpeningHours(f *testing.F) {
	f.Add(9*60, 17*60, 10*60)
	f.Fuzz(func(t *testing.T, opens, closes, visit int) {
		interval := place.Interval{Weekday: time.Monday, Opens: opens, Closes: closes}
		hours := place.OpeningHours{Known: true, Intervals: []place.Interval{interval}}
		if visit < 0 || visit > 22*60 {
			return
		}
		start := monday.Add(time.Duration(visit) * time.Minute)
		status := hours.OpenDuring(start, start.Add(time.Hour))
		switch status {
		case place.Open, place.Closed, place.UnknownStatus:
		default:
			t.Fatalf("invented status %q", status)
		}
		if status == place.Open && !interval.Valid() {
			t.Fatalf("an invalid interval reported open: %+v", interval)
		}
	})
}
