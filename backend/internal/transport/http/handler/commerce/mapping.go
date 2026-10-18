package commerce

import (
	"time"

	uuidlib "github.com/google/uuid"
	openapi "github.com/oapi-codegen/runtime/types"

	domain "github.com/SemiuAdesina/nuvro/backend/internal/domain/commerce"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/money"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/pricewatch"
	"github.com/SemiuAdesina/nuvro/backend/internal/transport/http/api"
	usecase "github.com/SemiuAdesina/nuvro/backend/internal/usecase/commerce"
)

func flightOffersOf(offers []domain.FlightOffer, now time.Time) []api.FlightOffer {
	values := make([]api.FlightOffer, 0, len(offers))
	for _, offer := range offers {
		value := api.FlightOffer{
			Id: offer.ID, Partner: offer.Partner, Product: api.FlightOfferProduct(offer.Product),
			Title: offer.Title, Price: moneyOf(offer.Price), Available: offer.Available,
			Freshness: api.OfferFreshness(offer.FreshnessAt(now)), ObservedAt: offer.ObservedAt.UTC(),
			Rank: offer.Rank, Stops: offer.Stops, DurationMinutes: offer.DurationMinutes,
		}
		value.Reasons = make([]api.FlightOfferReasons, 0, len(offer.Reasons))
		for _, reason := range offer.Reasons {
			value.Reasons = append(value.Reasons, api.FlightOfferReasons(reason))
		}
		value.OriginalPrice = optionalMoney(offer.OriginalPrice, offer.HasOriginal)
		value.ExpiresAt = instant(offer.ExpiresAt)
		value.CancellationSummary = optionalText(offer.CancellationSummary)
		value.Carrier = optionalText(offer.Carrier)
		value.LayoverMinutes = &offer.LayoverMinutes
		value.DepartAt = instant(offer.DepartAt)
		value.ArriveAt = instant(offer.ArriveAt)
		baggage := offer.BaggageIncluded
		changeable := offer.Changeable
		value.BaggageIncluded = &baggage
		value.Changeable = &changeable
		values = append(values, value)
	}
	return values
}

func stayOffersOf(offers []domain.StayOffer, now time.Time) []api.StayOffer {
	values := make([]api.StayOffer, 0, len(offers))
	for _, offer := range offers {
		value := api.StayOffer{
			Id: offer.ID, Partner: offer.Partner, Product: api.StayOfferProduct(offer.Product),
			Title: offer.Title, Price: moneyOf(offer.Price), Available: offer.Available,
			Freshness: api.OfferFreshness(offer.FreshnessAt(now)), ObservedAt: offer.ObservedAt.UTC(),
			Rank: offer.Rank, Nights: offer.Nights,
			Location: api.GeoPoint{
				Latitude: float32(offer.Location.Latitude), Longitude: float32(offer.Location.Longitude),
			},
		}
		value.Reasons = make([]api.StayOfferReasons, 0, len(offer.Reasons))
		for _, reason := range offer.Reasons {
			value.Reasons = append(value.Reasons, api.StayOfferReasons(reason))
		}
		rating := float32(offer.Rating)
		reviews := offer.ReviewCount
		cancellation := offer.FreeCancellation
		value.Rating = &rating
		value.ReviewCount = &reviews
		value.FreeCancellation = &cancellation
		value.Neighbourhood = optionalText(offer.Neighbourhood)
		value.OriginalPrice = optionalMoney(offer.OriginalPrice, offer.HasOriginal)
		value.ExpiresAt = instant(offer.ExpiresAt)
		value.CancellationSummary = optionalText(offer.CancellationSummary)
		values = append(values, value)
	}
	return values
}

func activityOffersOf(offers []domain.ActivityOffer, now time.Time) []api.ActivityOffer {
	values := make([]api.ActivityOffer, 0, len(offers))
	for _, offer := range offers {
		value := api.ActivityOffer{
			Id: offer.ID, Partner: offer.Partner, Product: api.ActivityOfferProduct(offer.Product),
			Title: offer.Title, Price: moneyOf(offer.Price), Available: offer.Available,
			Freshness: api.OfferFreshness(offer.FreshnessAt(now)), ObservedAt: offer.ObservedAt.UTC(),
			Rank: offer.Rank, DurationMinutes: offer.DurationMinutes,
			Location: api.GeoPoint{
				Latitude: float32(offer.Location.Latitude), Longitude: float32(offer.Location.Longitude),
			},
		}
		value.Reasons = make([]api.ActivityOfferReasons, 0, len(offer.Reasons))
		for _, reason := range offer.Reasons {
			value.Reasons = append(value.Reasons, api.ActivityOfferReasons(reason))
		}
		outdoor := offer.Outdoor
		rating := float32(offer.Rating)
		value.Outdoor = &outdoor
		value.Rating = &rating
		value.StartsAt = instant(offer.StartsAt)
		value.OriginalPrice = optionalMoney(offer.OriginalPrice, offer.HasOriginal)
		value.ExpiresAt = instant(offer.ExpiresAt)
		value.CancellationSummary = optionalText(offer.CancellationSummary)
		values = append(values, value)
	}
	return values
}

func statusOf(status usecase.SearchStatus) api.SearchStatus {
	partners := make([]api.PartnerStatus, 0, len(status.Partners))
	for _, partner := range status.Partners {
		value := api.PartnerStatus{Partner: partner.Partner, Available: partner.Available}
		value.Reason = optionalText(partner.Reason)
		partners = append(partners, value)
	}
	return api.SearchStatus{
		Partners: partners, RankingVersion: status.Ranking, SearchedAt: status.SearchedAt.UTC(),
	}
}

func disclosureOf(partners []usecase.PartnerStatus) api.CommerceDisclosure {
	name := "our booking partners"
	for _, partner := range partners {
		if partner.Available {
			name = partner.Partner
			break
		}
	}
	return api.CommerceDisclosure{Code: api.CommerceDisclosureCode(domain.EligibleBookingDisclosure), Partner: name}
}

func handoffOf(handoff domain.Handoff) api.BookingHandoff {
	created := handoff.CreatedAt.UTC()
	return api.BookingHandoff{
		Id: uuid(handoff.ID), OfferId: handoff.OfferID, Partner: handoff.Partner,
		Product: api.BookingHandoffProduct(handoff.Product), DestinationUrl: handoff.DestinationURL,
		Title: handoff.Title, Price: moneyOf(handoff.Price),
		Disclosure: api.CommerceDisclosure{
			Code: api.CommerceDisclosureCode(handoff.Disclosure.Code), Partner: handoff.Disclosure.Partner,
		},
		CreatedAt: &created, ExpiresAt: handoff.ExpiresAt.UTC(),
	}
}

func watchOf(watch pricewatch.Watch) api.PriceWatch {
	value := api.PriceWatch{
		Id: uuid(watch.ID), Product: api.PriceWatchProduct(watch.Product),
		Baseline: moneyOf(watch.Baseline), Status: api.PriceWatchStatus(watch.Status),
		CreatedAt: watch.CreatedAt.UTC(), ExpiresAt: watch.ExpiresAt.UTC(),
		Target: targetOf(watch.Target),
		Threshold: api.PriceWatchThreshold{
			PercentageDrop: &watch.Threshold.PercentageDrop, AbsoluteMinor: &watch.Threshold.AbsoluteMinor,
		},
	}
	if watch.TripID != "" {
		tripID := uuid(watch.TripID)
		value.TripId = &tripID
	}
	value.LastCheckedAt = instant(watch.LastCheckedAt)
	value.NextCheckAt = instant(watch.NextCheckAt)
	return value
}

func targetOf(target pricewatch.Target) api.PriceWatchTarget {
	value := api.PriceWatchTarget{
		StartDate: openapi.Date{Time: target.StartDate.UTC()},
		EndDate:   openapi.Date{Time: target.EndDate.UTC()},
		Adults:    target.Adults,
	}
	children := target.Children
	value.Children = &children
	value.Origin = optionalText(target.Origin)
	value.Destination = optionalText(target.Destination)
	value.PropertyId = optionalText(target.PropertyID)
	value.PlaceId = optionalText(target.PlaceID)
	if target.Cabin != "" {
		cabin := api.PriceWatchTargetCabin(target.Cabin)
		value.Cabin = &cabin
	}
	return value
}

func moneyOf(amount money.Amount) api.Money {
	return api.Money{AmountMinor: amount.Minor(), Currency: amount.Currency().String()}
}

func optionalMoney(amount money.Amount, present bool) *api.Money {
	if !present {
		return nil
	}
	value := moneyOf(amount)
	return &value
}

func optionalText(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func instant(value time.Time) *time.Time {
	if value.IsZero() {
		return nil
	}
	moment := value.UTC()
	return &moment
}

func uuid(value string) openapi.UUID {
	parsed, err := uuidlib.Parse(value)
	if err != nil {
		return openapi.UUID{}
	}
	return parsed
}
