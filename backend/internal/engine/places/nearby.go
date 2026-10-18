package places

import (
	"slices"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/geo"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/place"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/route"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/weather"
)

type Context struct {
	Now              time.Time
	Origin           geo.Point
	AvailableMinutes int
	Interests        []string
	Weather          weather.Window
	Routes           map[string]route.Evidence
}

type Ranked struct {
	Place         place.Place
	Score         int
	Reasons       []string
	Fits          bool
	TravelMinutes int
	WeatherImpact weather.Impact
}

const maximumScore = 100

func Rank(candidates []place.Place, context Context) []Ranked {
	ranked := make([]Ranked, 0, len(candidates))
	for _, candidate := range candidates {
		ranked = append(ranked, evaluate(candidate, context))
	}
	slices.SortStableFunc(ranked, func(left, right Ranked) int {
		if left.Fits != right.Fits {
			if left.Fits {
				return -1
			}
			return 1
		}
		if order := right.Score - left.Score; order != 0 {
			return order
		}
		if order := left.TravelMinutes - right.TravelMinutes; order != 0 {
			return order
		}
		return slices.Compare([]string{left.Place.ID}, []string{right.Place.ID})
	})
	return ranked
}

func evaluate(candidate place.Place, context Context) Ranked {
	result := Ranked{Place: candidate, TravelMinutes: -1, WeatherImpact: weather.UnknownImpact, Fits: true}
	score := 0
	score += interestScore(candidate, context.Interests, &result)
	score += openScore(candidate, &result)
	score += ratingScore(candidate)
	travel, known := travelMinutes(candidate, context)
	if known {
		result.TravelMinutes = travel
		score += proximityScore(travel, &result)
	}
	result.WeatherImpact = weatherImpact(candidate, context)
	score += weatherScore(result.WeatherImpact, &result)
	result.Fits = fitsAvailableTime(candidate, travel, known, context, &result)
	result.Score = min(score, maximumScore)
	return result
}

func interestScore(candidate place.Place, interests []string, result *Ranked) int {
	for _, interest := range interests {
		if matchesInterest(candidate.Category, interest) {
			result.Reasons = append(result.Reasons, "interest_fit")
			return 30
		}
	}
	return 0
}

var categoriesByInterest = map[string][]place.Category{
	"food":      {place.Restaurant, place.Cafe, place.Market},
	"culture":   {place.Museum, place.Attraction},
	"history":   {place.Museum, place.Attraction},
	"nature":    {place.Park, place.Viewpoint, place.Beach},
	"beach":     {place.Beach},
	"shopping":  {place.Shopping, place.Market},
	"nightlife": {place.Nightlife},
	"adventure": {place.Experience},
}

func matchesInterest(category place.Category, interest string) bool {
	return slices.Contains(categoriesByInterest[interest], category)
}

func openScore(candidate place.Place, result *Ranked) int {
	switch candidate.Status {
	case place.Open:
		result.Reasons = append(result.Reasons, "open_now")
		return 25
	case place.Closed:
		result.Reasons = append(result.Reasons, "closed_now")
		return 0
	default:
		result.Reasons = append(result.Reasons, "opening_hours_unknown")
		return 10
	}
}

func ratingScore(candidate place.Place) int {
	if candidate.ReviewCount < 10 || candidate.Rating <= 0 {
		return 5
	}
	return int(candidate.Rating * 3)
}

func proximityScore(travelMinutes int, result *Ranked) int {
	switch {
	case travelMinutes <= 10:
		result.Reasons = append(result.Reasons, "very_close")
		return 25
	case travelMinutes <= 25:
		result.Reasons = append(result.Reasons, "nearby")
		return 15
	default:
		return 5
	}
}

func weatherImpact(candidate place.Place, context Context) weather.Impact {
	if !context.Weather.Known() {
		return weather.UnknownImpact
	}
	observation, found := context.Weather.HourAt(context.Now)
	if !found {
		if context.Weather.Current == nil {
			return weather.UnknownImpact
		}
		observation = *context.Weather.Current
	}
	return weather.Assess(observation, weather.SensitivityOf(candidate.Category))
}

func weatherScore(impact weather.Impact, result *Ranked) int {
	switch impact {
	case weather.NoImpact:
		return 20
	case weather.LowImpact:
		return 15
	case weather.ModerateImpact:
		result.Reasons = append(result.Reasons, "weather_may_affect")
		return 5
	case weather.HighImpact:
		result.Reasons = append(result.Reasons, "weather_unsuitable")
		return 0
	default:
		return 10
	}
}

func travelMinutes(candidate place.Place, context Context) (int, bool) {
	if evidence, found := context.Routes[candidate.ID]; found && evidence.Known() {
		return int(evidence.Duration().Minutes()), true
	}
	if context.Origin.IsZero() {
		return 0, false
	}
	return 0, false
}

func fitsAvailableTime(candidate place.Place, travel int, travelKnown bool, context Context, result *Ranked) bool {
	if context.AvailableMinutes <= 0 {
		return true
	}
	required := int(candidate.Duration().Minutes())
	if travelKnown {
		required += travel * 2
	}
	if required > context.AvailableMinutes {
		result.Reasons = append(result.Reasons, "does_not_fit_available_time")
		return false
	}
	result.Reasons = append(result.Reasons, "fits_available_time")
	return true
}
