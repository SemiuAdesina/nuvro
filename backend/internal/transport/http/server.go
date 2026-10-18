package http

import (
	"context"

	"github.com/SemiuAdesina/nuvro/backend/internal/transport/http/api"
)

type systemOperations interface {
	GetLiveness(context.Context, api.GetLivenessRequestObject) (api.GetLivenessResponseObject, error)
	GetReadiness(context.Context, api.GetReadinessRequestObject) (api.GetReadinessResponseObject, error)
}

type authOperations interface {
	CreateSession(context.Context, api.CreateSessionRequestObject) (api.CreateSessionResponseObject, error)
	RefreshSession(context.Context, api.RefreshSessionRequestObject) (api.RefreshSessionResponseObject, error)
	DeleteSession(context.Context, api.DeleteSessionRequestObject) (api.DeleteSessionResponseObject, error)
}

type profileOperations interface {
	GetMe(context.Context, api.GetMeRequestObject) (api.GetMeResponseObject, error)
	DeleteMe(context.Context, api.DeleteMeRequestObject) (api.DeleteMeResponseObject, error)
	PutAppleIdentity(context.Context, api.PutAppleIdentityRequestObject) (api.PutAppleIdentityResponseObject, error)
	PatchMe(context.Context, api.PatchMeRequestObject) (api.PatchMeResponseObject, error)
	GetPreferences(context.Context, api.GetPreferencesRequestObject) (api.GetPreferencesResponseObject, error)
	PatchPreferences(context.Context, api.PatchPreferencesRequestObject) (api.PatchPreferencesResponseObject, error)
}

type intelligenceOperations interface {
	GetDestination(context.Context, api.GetDestinationRequestObject) (api.GetDestinationResponseObject, error)
	GetDestinationMedia(context.Context, api.GetDestinationMediaRequestObject) (api.GetDestinationMediaResponseObject, error)
}

type journeyOperations interface {
	GenerateItinerary(context.Context, api.GenerateItineraryRequestObject) (api.GenerateItineraryResponseObject, error)
	GetItinerary(context.Context, api.GetItineraryRequestObject) (api.GetItineraryResponseObject, error)
	PatchItineraryItem(context.Context, api.PatchItineraryItemRequestObject) (api.PatchItineraryItemResponseObject, error)
	GetTripBudget(context.Context, api.GetTripBudgetRequestObject) (api.GetTripBudgetResponseObject, error)
	SearchNearby(context.Context, api.SearchNearbyRequestObject) (api.SearchNearbyResponseObject, error)
}

type runtimeOperations interface {
	GetActiveTrip(context.Context, api.GetActiveTripRequestObject) (api.GetActiveTripResponseObject, error)
	ProposeReplan(context.Context, api.ProposeReplanRequestObject) (api.ProposeReplanResponseObject, error)
	ApplyReplan(context.Context, api.ApplyReplanRequestObject) (api.ApplyReplanResponseObject, error)
	ListTripAlerts(context.Context, api.ListTripAlertsRequestObject) (api.ListTripAlertsResponseObject, error)
	PatchTripAlert(context.Context, api.PatchTripAlertRequestObject) (api.PatchTripAlertResponseObject, error)
	TranslateForTrip(context.Context, api.TranslateForTripRequestObject) (api.TranslateForTripResponseObject, error)
	GetTripPhrases(context.Context, api.GetTripPhrasesRequestObject) (api.GetTripPhrasesResponseObject, error)
	GetOfflinePack(context.Context, api.GetOfflinePackRequestObject) (api.GetOfflinePackResponseObject, error)
}

type personalizationOperations interface {
	GetPersonalization(context.Context, api.GetPersonalizationRequestObject) (api.GetPersonalizationResponseObject, error)
	PatchPersonalization(context.Context, api.PatchPersonalizationRequestObject) (api.PatchPersonalizationResponseObject, error)
	ClearPersonalization(context.Context, api.ClearPersonalizationRequestObject) (api.ClearPersonalizationResponseObject, error)
	ListPersonalizationInsights(context.Context, api.ListPersonalizationInsightsRequestObject) (api.ListPersonalizationInsightsResponseObject, error)
	PatchPersonalizationInsight(context.Context, api.PatchPersonalizationInsightRequestObject) (api.PatchPersonalizationInsightResponseObject, error)
	SubmitTripFeedback(context.Context, api.SubmitTripFeedbackRequestObject) (api.SubmitTripFeedbackResponseObject, error)
	GetTripFeedback(context.Context, api.GetTripFeedbackRequestObject) (api.GetTripFeedbackResponseObject, error)
}

type commerceOperations interface {
	SearchFlightOffers(context.Context, api.SearchFlightOffersRequestObject) (api.SearchFlightOffersResponseObject, error)
	SearchStayOffers(context.Context, api.SearchStayOffersRequestObject) (api.SearchStayOffersResponseObject, error)
	SearchActivityOffers(context.Context, api.SearchActivityOffersRequestObject) (api.SearchActivityOffersResponseObject, error)
	CreateBookingHandoff(context.Context, api.CreateBookingHandoffRequestObject) (api.CreateBookingHandoffResponseObject, error)
	ListPriceWatches(context.Context, api.ListPriceWatchesRequestObject) (api.ListPriceWatchesResponseObject, error)
	CreatePriceWatch(context.Context, api.CreatePriceWatchRequestObject) (api.CreatePriceWatchResponseObject, error)
	DeletePriceWatch(context.Context, api.DeletePriceWatchRequestObject) (api.DeletePriceWatchResponseObject, error)
}

type discoverOperations interface {
	DiscoverMatches(context.Context, api.DiscoverMatchesRequestObject) (api.DiscoverMatchesResponseObject, error)
}

type tripOperations interface {
	CreateTrip(context.Context, api.CreateTripRequestObject) (api.CreateTripResponseObject, error)
	ListTrips(context.Context, api.ListTripsRequestObject) (api.ListTripsResponseObject, error)
	GetTrip(context.Context, api.GetTripRequestObject) (api.GetTripResponseObject, error)
	PatchTrip(context.Context, api.PatchTripRequestObject) (api.PatchTripResponseObject, error)
}

type server struct {
	systemOperations
	profileOperations
	authOperations
	tripOperations
	discoverOperations
	intelligenceOperations
	journeyOperations
	runtimeOperations
	commerceOperations
	personalizationOperations
}
