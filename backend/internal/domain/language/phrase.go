package language

import "slices"

type Category string

const (
	Hotel      Category = "hotel"
	Restaurant Category = "restaurant"
	Transport  Category = "transport"
	Directions Category = "directions"
	Shopping   Category = "shopping"
	Emergency  Category = "emergency"
	Medical    Category = "medical"
	General    Category = "general"
)

func Categories() []Category {
	return []Category{Hotel, Restaurant, Transport, Directions, Shopping, Emergency, Medical, General}
}

func (c Category) Valid() bool { return slices.Contains(Categories(), c) }

func ParseCategory(raw string) (Category, bool) {
	candidate := Category(raw)
	return candidate, candidate.Valid()
}

type Phrase struct {
	Key           string
	Category      Category
	SourceText    string
	TargetText    string
	HasTargetText bool
}

type Pack struct {
	Source  Code
	Target  Code
	Phrases []Phrase
}

func (p Pack) ByCategory(category Category) []Phrase {
	matching := make([]Phrase, 0, len(p.Phrases))
	for _, phrase := range p.Phrases {
		if phrase.Category == category {
			matching = append(matching, phrase)
		}
	}
	return matching
}

type Translation struct {
	Source     Code
	Target     Code
	SourceText string
	Translated string
	Detected   bool
	Confidence string
	Provider   string
}
