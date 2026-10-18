package place

import "slices"

type Category string

const (
	Restaurant Category = "restaurant"
	Cafe       Category = "cafe"
	Attraction Category = "attraction"
	Museum     Category = "museum"
	Park       Category = "park"
	Beach      Category = "beach"
	Market     Category = "market"
	Nightlife  Category = "nightlife"
	Shopping   Category = "shopping"
	Viewpoint  Category = "viewpoint"
	Transport  Category = "transport"
	Hotel      Category = "hotel"
	Experience Category = "experience"
	Other      Category = "other"
)

func Categories() []Category {
	return []Category{
		Restaurant, Cafe, Attraction, Museum, Park, Beach, Market,
		Nightlife, Shopping, Viewpoint, Transport, Hotel, Experience, Other,
	}
}

func (c Category) Valid() bool { return slices.Contains(Categories(), c) }

func (c Category) Outdoor() bool {
	switch c {
	case Beach, Park, Viewpoint, Market:
		return true
	default:
		return false
	}
}

func (c Category) Indoor() bool {
	switch c {
	case Museum, Restaurant, Cafe, Shopping, Hotel:
		return true
	default:
		return false
	}
}

func ParseCategory(raw string) Category {
	candidate := Category(raw)
	if candidate.Valid() {
		return candidate
	}
	return Other
}
