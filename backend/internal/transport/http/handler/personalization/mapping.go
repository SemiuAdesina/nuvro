package personalization

import (
	"time"

	uuidlib "github.com/google/uuid"
	openapi "github.com/oapi-codegen/runtime/types"

	feedbackdomain "github.com/SemiuAdesina/nuvro/backend/internal/domain/feedback"
	domain "github.com/SemiuAdesina/nuvro/backend/internal/domain/personalization"
	"github.com/SemiuAdesina/nuvro/backend/internal/transport/http/api"
)

func settingsOf(value domain.Context) api.PersonalizationSettings {
	suppressed := make([]string, 0, len(value.Suppressed))
	for _, key := range value.Suppressed {
		suppressed = append(suppressed, string(key))
	}
	return api.PersonalizationSettings{
		HistoryLearningEnabled: value.HistoryLearningEnabled,
		RulesVersion:           value.RulesVersion,
		GeneratedAt:            value.GeneratedAt.UTC(),
		Explicit:               preferencesOf(value.Explicit),
		Learned:                preferencesOf(value.Learned),
		Suppressed:             suppressed,
	}
}

func preferencesOf(values []domain.Preference) []api.PersonalizationPreference {
	preferences := make([]api.PersonalizationPreference, 0, len(values))
	for _, value := range values {
		preferences = append(preferences, api.PersonalizationPreference{
			Key: string(value.Key), Origin: api.PreferenceOrigin(value.Origin),
			Confidence: api.PersonalizationPreferenceConfidence(value.Confidence),
		})
	}
	return preferences
}

func insightsOf(values []domain.Insight) []api.PersonalizationInsight {
	insights := make([]api.PersonalizationInsight, 0, len(values))
	for _, value := range values {
		insights = append(insights, insightOf(value))
	}
	return insights
}

func insightOf(value domain.Insight) api.PersonalizationInsight {
	insight := api.PersonalizationInsight{
		Id: uuid(value.ID), Key: string(value.Key), Category: value.Category,
		SummaryCode: value.SummaryCode, Confidence: api.PersonalizationInsightConfidence(value.Confidence),
		EvidenceCount: value.EvidenceCount, Status: api.PersonalizationInsightStatus(value.Status),
		Editable: value.Editable,
	}
	if len(value.Parameters) > 0 {
		parameters := value.Parameters
		insight.Parameters = &parameters
	}
	insight.CreatedAt = instant(value.CreatedAt)
	insight.UpdatedAt = instant(value.UpdatedAt)
	return insight
}

func feedbackOf(value feedbackdomain.TripFeedback) api.TripFeedback {
	result := api.TripFeedback{
		TripId: uuid(value.TripID), SubmittedAt: value.SubmittedAt.UTC(),
		Positives: aspectsOf(value.Positives), Negatives: aspectsOf(value.Negatives),
	}
	if value.Pace != "" {
		pace := api.TripFeedbackPace(value.Pace)
		result.Pace = &pace
	}
	if value.TravelBurden != "" {
		burden := api.TripFeedbackTravelBurden(value.TravelBurden)
		result.TravelBurden = &burden
	}
	if value.ReturnIntent != "" {
		intent := api.TripFeedbackReturnIntent(value.ReturnIntent)
		result.ReturnIntent = &intent
	}
	result.UpdatedAt = instant(value.UpdatedAt)
	return result
}

func aspectsOf(values []feedbackdomain.Aspect) []api.FeedbackAspect {
	aspects := make([]api.FeedbackAspect, 0, len(values))
	for _, value := range values {
		aspects = append(aspects, api.FeedbackAspect(value))
	}
	return aspects
}

func aspectCodes(values *[]api.FeedbackAspect) []string {
	if values == nil {
		return nil
	}
	codes := make([]string, 0, len(*values))
	for _, value := range *values {
		codes = append(codes, string(value))
	}
	return codes
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
