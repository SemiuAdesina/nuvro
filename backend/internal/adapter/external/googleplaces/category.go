package googleplaces

import "github.com/SemiuAdesina/nuvro/backend/internal/domain/place"

var categoryByType = map[string]place.Category{
	"restaurant": place.Restaurant, "meal_takeaway": place.Restaurant, "bakery": place.Restaurant,
	"cafe": place.Cafe, "coffee_shop": place.Cafe,
	"tourist_attraction": place.Attraction, "historical_landmark": place.Attraction,
	"church": place.Attraction, "monument": place.Attraction,
	"museum": place.Museum, "art_gallery": place.Museum,
	"park": place.Park, "national_park": place.Park, "botanical_garden": place.Park, "garden": place.Park,
	"beach":  place.Beach,
	"market": place.Market, "farmers_market": place.Market,
	"bar": place.Nightlife, "night_club": place.Nightlife,
	"shopping_mall": place.Shopping, "store": place.Shopping, "clothing_store": place.Shopping,
	"observation_deck": place.Viewpoint, "scenic_point": place.Viewpoint,
	"airport": place.Transport, "train_station": place.Transport, "bus_station": place.Transport,
	"subway_station": place.Transport, "transit_station": place.Transport,
	"hotel": place.Hotel, "lodging": place.Hotel, "resort_hotel": place.Hotel,
	"amusement_park": place.Experience, "aquarium": place.Experience, "zoo": place.Experience,
	"tourist_information_center": place.Attraction,
}

var typeByCategory = map[place.Category][]string{
	place.Restaurant: {"restaurant"},
	place.Cafe:       {"cafe"},
	place.Attraction: {"tourist_attraction"},
	place.Museum:     {"museum", "art_gallery"},
	place.Park:       {"park"},
	place.Beach:      {"beach"},
	place.Market:     {"market"},
	place.Nightlife:  {"bar", "night_club"},
	place.Shopping:   {"shopping_mall"},
	place.Viewpoint:  {"observation_deck"},
	place.Transport:  {"transit_station"},
	place.Hotel:      {"hotel"},
	place.Experience: {"amusement_park", "aquarium", "zoo"},
}

func category(primary string, types []string) place.Category {
	if mapped, found := categoryByType[primary]; found {
		return mapped
	}
	for _, value := range types {
		if mapped, found := categoryByType[value]; found {
			return mapped
		}
	}
	return place.Other
}

func includedTypes(categories []place.Category) []string {
	seen := map[string]struct{}{}
	included := make([]string, 0, len(categories))
	for _, wanted := range categories {
		for _, value := range typeByCategory[wanted] {
			if _, duplicate := seen[value]; duplicate {
				continue
			}
			seen[value] = struct{}{}
			included = append(included, value)
		}
	}
	return included
}
