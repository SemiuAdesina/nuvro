package commerce

import (
	"context"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/alert"
	domain "github.com/SemiuAdesina/nuvro/backend/internal/domain/commerce"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/destination"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/itinerary"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/pricewatch"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/trip"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
	engine "github.com/SemiuAdesina/nuvro/backend/internal/engine/commerce"
)

type Trips interface {
	Read(context.Context, user.Principal, string) (trip.Trip, error)
}

type Catalog interface {
	Destination(context.Context, string) (destination.Destination, error)
}

type Itineraries interface {
	Itinerary(context.Context, string) (itinerary.Itinerary, error)
}

type HandoffStore interface {
	SaveHandoff(context.Context, domain.Handoff) error
	Handoff(context.Context, string) (domain.Handoff, error)
	RecordClick(context.Context, domain.Click) error
}

type ConversionStore interface {
	ClaimWebhookEvent(ctx context.Context, eventID, partner, providerEventID, eventType, status string) (bool, error)
	MarkWebhookProcessed(ctx context.Context, partner, providerEventID string, at time.Time) error
	SaveConversion(context.Context, domain.Conversion) (domain.Conversion, error)
	Conversion(ctx context.Context, partner, providerConversionID string) (domain.Conversion, error)
}

type WatchStore interface {
	SavePriceWatch(context.Context, pricewatch.Watch) error
	PriceWatches(context.Context, string) ([]pricewatch.Watch, error)
	PriceWatch(context.Context, string) (pricewatch.Watch, error)
	ActivePriceWatchCount(context.Context, string) (int64, error)
	EndPriceWatch(context.Context, string, pricewatch.Status) error
}

type Entitlements interface {
	AllowsPriceMonitoring(context.Context, user.Principal) (bool, error)
}

type Alerts interface {
	Record(context.Context, []alert.Alert) ([]alert.Alert, error)
}

type Identifiers interface{ Generate() (string, error) }

type Providers struct {
	Flights    []engine.FlightOfferProvider
	Stays      []engine.StayOfferProvider
	Activities []engine.ActivityOfferProvider
	Transfers  []engine.TransferOfferProvider
}
