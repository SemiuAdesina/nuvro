package alerting

import (
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/alert"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/itinerary"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/travel"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/weather"
)

type Input struct {
	TripID    string
	Timezone  string
	Now       time.Time
	Itinerary itinerary.Itinerary
	Weather   weather.Window
	Departure travel.Departure
	NextItem  *itinerary.Item
}

type Evaluator interface {
	Name() string
	Evaluate(Input) []alert.Alert
}

func window(moment time.Time) string { return moment.UTC().Format("2006-01-02T15") }

func base(input Input, kind alert.Type, severity alert.Severity, code string, resource string, action alert.Action) alert.Alert {
	return alert.Alert{
		TripID: input.TripID, Type: kind, Severity: severity, TitleCode: code,
		RelatedItemID: resource, Action: action, Status: alert.Unread,
		Fingerprint: alert.Fingerprint(input.TripID, kind, resource, window(input.Now)),
		CreatedAt:   input.Now.UTC(), EffectiveAt: input.Now.UTC(),
	}
}
