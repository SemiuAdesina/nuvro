package discover

import (
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/discovery"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/money"
	"github.com/SemiuAdesina/nuvro/backend/internal/engine/understanding"
	"github.com/SemiuAdesina/nuvro/backend/internal/transport/http/api"
	"github.com/SemiuAdesina/nuvro/backend/internal/usecase/discover"
)

func discoveryRequest(body api.MatchRequest) (discover.Request, bool) {
	budget, err := money.New(body.Budget.AmountMinor, money.Currency(body.Budget.Currency))
	if err != nil {
		return discover.Request{}, false
	}
	supplied := understanding.Supplied{
		Budget:            &budget,
		Interests:         values(body.Interests),
		Weather:           enumText(body.Weather),
		Crowds:            enumText(body.Crowds),
		Pace:              enumText(body.Pace),
		FlightTolerance:   enumText(body.FlightTolerance),
		Paperwork:         enumText(body.Paperwork),
		MaxFlightMinutes:  body.MaxFlightMinutes,
		DepartureAirport:  text(body.DepartureAirport),
		ResidenceCountry:  text(body.ResidenceCountry),
		PassportCountries: values(body.PassportCountries),
		ResidencePermits:  values(body.ResidencePermits),
	}
	return discover.Request{
		Overrides: supplied,
		Party:     discovery.Party{Adults: body.Adults, Children: count(body.Children)},
		Window:    discovery.Window{Start: calendar(body.StartDate.Time), End: calendar(body.EndDate.Time)},
		Query:     text(body.Query),
		Limit:     count(body.Limit),
	}, true
}

func calendar(value time.Time) time.Time {
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
}

func values(source *[]string) []string {
	if source == nil {
		return nil
	}
	return *source
}

func text(source *string) string {
	if source == nil {
		return ""
	}
	return *source
}

func count(source *int) int {
	if source == nil {
		return 0
	}
	return *source
}

func enumText[T ~string](source *T) string {
	if source == nil {
		return ""
	}
	return string(*source)
}
