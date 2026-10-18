package commerce

import (
	"context"
	"time"

	domain "github.com/SemiuAdesina/nuvro/backend/internal/domain/commerce"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/geo"
)

type FlightQuery struct {
	Origin      string
	Destination string
	DepartDate  time.Time
	ReturnDate  time.Time
	Adults      int
	Children    int
	Cabin       string
	Currency    string
}

type StayQuery struct {
	Centre       geo.Point
	CheckIn      time.Time
	CheckOut     time.Time
	Adults       int
	Children     int
	Currency     string
	RadiusMeters int
}

type ActivityQuery struct {
	Centre    geo.Point
	Date      time.Time
	Adults    int
	Children  int
	Currency  string
	Interests []string
}

type TransferQuery struct {
	Origin      geo.Point
	Destination geo.Point
	PickupAt    time.Time
	Passengers  int
	Currency    string
}

type Priced struct {
	Commercials domain.Commercials
}

type FlightOfferProvider interface {
	Partner() domain.Partner
	Flights(context.Context, FlightQuery) ([]domain.FlightOffer, []domain.Commercials, error)
}

type StayOfferProvider interface {
	Partner() domain.Partner
	Stays(context.Context, StayQuery) ([]domain.StayOffer, []domain.Commercials, error)
}

type ActivityOfferProvider interface {
	Partner() domain.Partner
	Activities(context.Context, ActivityQuery) ([]domain.ActivityOffer, []domain.Commercials, error)
}

type TransferOfferProvider interface {
	Partner() domain.Partner
	Transfers(context.Context, TransferQuery) ([]domain.TransferOffer, []domain.Commercials, error)
}

type OfferStore interface {
	SaveOffers(ctx context.Context, offers []StoredOffer, ttl time.Duration) error
	Offer(ctx context.Context, offerID string) (StoredOffer, bool)
}

type StoredOffer struct {
	Offer       domain.Offer
	Commercials domain.Commercials
	Destination string
}

type AttributionPublisher interface {
	Publish(ctx context.Context, handoff domain.Handoff, commercials domain.Commercials) error
}
