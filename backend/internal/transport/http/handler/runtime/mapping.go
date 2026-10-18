package runtime

import (
	uuidlib "github.com/google/uuid"
	openapi "github.com/oapi-codegen/runtime/types"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/alert"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/evidence"
	domainlanguage "github.com/SemiuAdesina/nuvro/backend/internal/domain/language"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/replan"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/route"
	domain "github.com/SemiuAdesina/nuvro/backend/internal/domain/travel"
	"github.com/SemiuAdesina/nuvro/backend/internal/transport/http/api"
	"github.com/SemiuAdesina/nuvro/backend/internal/transport/http/handler/journey"
)

func departureOf(departure domain.Departure) api.Departure {
	value := api.Departure{Status: api.DepartureStatus(departure.Status)}
	if departure.HasLeaveBy {
		leaveBy := departure.LeaveBy.UTC()
		minutes := departure.MinutesUntilLeave
		value.LeaveBy = &leaveBy
		value.MinutesUntilLeave = &minutes
	}
	if departure.ArrivalBuffer > 0 {
		buffer := int(departure.ArrivalBuffer.Minutes())
		value.ArrivalBufferMinutes = &buffer
	}
	if departure.HasRoute {
		value.Route = routeOf(departure.Route)
	}
	return value
}

func routeOf(evidence route.Evidence) *api.RouteSummary {
	summary := api.RouteSummary{
		Status: api.RouteSummaryStatus(evidence.Status), Mode: api.RouteSummaryMode(evidence.Mode),
	}
	if evidence.Confidence != "" {
		confidence := api.RouteSummaryConfidence(evidence.Confidence)
		summary.Confidence = &confidence
	}
	if evidence.Known() {
		distance := evidence.DistanceMeters
		duration := evidence.DurationSeconds
		summary.DistanceMeters = &distance
		summary.DurationSeconds = &duration
	}
	return &summary
}

func actionsOf(actions []domain.Action) []api.QuickAction {
	values := make([]api.QuickAction, 0, len(actions))
	for _, action := range actions {
		value := api.QuickAction{Type: api.QuickActionType(action.Type), Enabled: action.Enabled}
		if action.ResourceID != "" {
			resource := action.ResourceID
			value.ResourceId = &resource
		}
		values = append(values, value)
	}
	return values
}

func alertsOf(values []alert.Alert) []api.TripAlert {
	alerts := make([]api.TripAlert, 0, len(values))
	for _, value := range values {
		alerts = append(alerts, alertOf(value))
	}
	return alerts
}

func alertOf(value alert.Alert) api.TripAlert {
	result := api.TripAlert{
		Id: uuid(value.ID), TripId: uuid(value.TripID), Type: api.TripAlertType(value.Type),
		Severity: api.TripAlertSeverity(value.Severity), TitleCode: value.TitleCode,
		Action: api.TripAlertAction(value.Action), Status: api.TripAlertStatus(value.Status),
		EffectiveAt: value.EffectiveAt.UTC(),
	}
	if len(value.MessageData) > 0 {
		data := value.MessageData
		result.MessageData = &data
	}
	if value.RelatedItemID != "" {
		related := value.RelatedItemID
		result.RelatedItemId = &related
	}
	if !value.ExpiresAt.IsZero() {
		expires := value.ExpiresAt.UTC()
		result.ExpiresAt = &expires
	}
	result.Freshness = freshnessOf(value.Freshness)
	return result
}

func freshnessOf(freshness evidence.Freshness) *api.Freshness {
	if freshness.ObservedAt.IsZero() {
		return nil
	}
	return &api.Freshness{
		ObservedAt: freshness.ObservedAt.UTC(), ValidUntil: freshness.ValidUntil.UTC(),
		Status: api.FreshnessStatus(statusOrCached(freshness.Status)),
	}
}

func statusOrCached(status evidence.Status) evidence.Status {
	if status == "" {
		return evidence.Cached
	}
	return status
}

func proposalOf(proposal replan.Proposal) api.ReplanProposal {
	changes := make([]api.ReplanChange, 0, len(proposal.Changes))
	for _, change := range proposal.Changes {
		changes = append(changes, changeOf(change))
	}
	value := api.ReplanProposal{
		TripId: uuid(proposal.TripID), Trigger: api.ReplanProposalTrigger(proposal.Trigger),
		ItineraryVersion: proposal.ItineraryVersion, Changes: changes,
		Reasons: reasonsOf(proposal.Reasons), PreservedConstraints: constraintsOf(proposal.Preserved),
		Valid: proposal.Valid(), CreatedAt: proposal.CreatedAt.UTC(), ExpiresAt: proposal.ExpiresAt.UTC(),
	}
	if proposal.ID != "" {
		identifier := uuid(proposal.ID)
		value.Id = &identifier
	}
	if proposal.Summary != "" {
		summary := proposal.Summary
		value.Summary = &summary
	}
	if proposal.HasBudgetDelta {
		value.BudgetDelta = &api.Money{
			AmountMinor: proposal.BudgetDelta.Minor(), Currency: proposal.BudgetDelta.Currency().String(),
		}
	}
	if len(proposal.Violations) > 0 {
		violations := journey.ViolationsOf(proposal.Violations)
		value.Violations = &violations
	}
	return value
}

func changeOf(change replan.Change) api.ReplanChange {
	value := api.ReplanChange{Kind: api.ReplanChangeKind(change.Kind), Reasons: changeReasons(change.Reasons)}
	if change.Original != nil {
		original := journey.ItemOf(*change.Original)
		value.Original = &original
	}
	if change.Proposed != nil {
		proposed := journey.ItemOf(*change.Proposed)
		value.Proposed = &proposed
	}
	return value
}

func changeReasons(reasons []replan.Reason) []api.ReplanChangeReasons {
	values := make([]api.ReplanChangeReasons, 0, len(reasons))
	for _, reason := range reasons {
		values = append(values, api.ReplanChangeReasons(reason))
	}
	return values
}

func reasonsOf(reasons []replan.Reason) []string {
	values := make([]string, 0, len(reasons))
	for _, reason := range reasons {
		values = append(values, string(reason))
	}
	return values
}

func constraintsOf(constraints []replan.Constraint) []api.ReplanProposalPreservedConstraints {
	values := make([]api.ReplanProposalPreservedConstraints, 0, len(constraints))
	for _, constraint := range constraints {
		values = append(values, api.ReplanProposalPreservedConstraints(constraint))
	}
	return values
}

func translationOf(translation domainlanguage.Translation) api.Translation {
	value := api.Translation{
		SourceLanguage: translation.Source.String(), TargetLanguage: translation.Target.String(),
		SourceText: translation.SourceText, TranslatedText: translation.Translated,
		Detected: translation.Detected, Confidence: api.TranslationConfidence(translation.Confidence),
	}
	if translation.Provider != "" {
		provider := translation.Provider
		value.Provider = &provider
	}
	return value
}

func packOf(pack domainlanguage.Pack) api.PhrasePack {
	return api.PhrasePack{
		SourceLanguage: pack.Source.String(), TargetLanguage: pack.Target.String(),
		Phrases: phrasesOf(pack.Phrases),
	}
}

func phrasesOf(phrases []domainlanguage.Phrase) []api.Phrase {
	values := make([]api.Phrase, 0, len(phrases))
	for _, phrase := range phrases {
		value := api.Phrase{
			Key: phrase.Key, Category: api.PhraseCategory(phrase.Category),
			SourceText: phrase.SourceText, Translated: phrase.HasTargetText,
		}
		if phrase.HasTargetText {
			target := phrase.TargetText
			value.TargetText = &target
		}
		values = append(values, value)
	}
	return values
}

func uuid(value string) openapi.UUID {
	parsed, err := uuidlib.Parse(value)
	if err != nil {
		return openapi.UUID{}
	}
	return parsed
}
