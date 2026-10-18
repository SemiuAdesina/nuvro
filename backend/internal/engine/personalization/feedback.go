package personalization

import (
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/feedback"
	domain "github.com/SemiuAdesina/nuvro/backend/internal/domain/personalization"
)

var subjectByAspect = map[feedback.Aspect]struct {
	Subject domain.SubjectType
	Value   string
}{
	feedback.CoastalPlaces:     {domain.Interest, "coastal"},
	feedback.LocalFood:         {domain.Interest, "food"},
	feedback.CultureAndHistory: {domain.Interest, "culture"},
	feedback.NatureAndOutdoors: {domain.Interest, "nature"},
	feedback.Nightlife:         {domain.Interest, "nightlife"},
	feedback.QuietArea:         {domain.StayArea, "quiet"},
	feedback.CentralArea:       {domain.StayArea, "central"},
	feedback.LongTransfers:     {domain.TravelBurden, "long_transfers"},
	feedback.BusySchedule:      {domain.TripPace, "packed"},
}

func SignalsFromFeedback(submission feedback.TripFeedback) []domain.Signal {
	signals := make([]domain.Signal, 0)
	for _, aspect := range submission.Positives {
		signals = append(signals, aspectSignal(submission, aspect, domain.FeedbackPositive)...)
	}
	for _, aspect := range submission.Negatives {
		signals = append(signals, negativeAspectSignal(submission, aspect)...)
	}
	signals = append(signals, paceSignals(submission)...)
	signals = append(signals, burdenSignals(submission)...)
	return signals
}

func aspectSignal(submission feedback.TripFeedback, aspect feedback.Aspect, signalType domain.SignalType) []domain.Signal {
	mapping, known := subjectByAspect[aspect]
	if !known {
		return nil
	}
	signal := domain.NewSignal(submission.UserID, signalType, mapping.Subject, mapping.Value, submission.SubmittedAt)
	signal.TripID = submission.TripID
	signal.SourceEvent = "feedback:" + submission.TripID + ":" + string(aspect)
	return []domain.Signal{signal}
}

func negativeAspectSignal(submission feedback.TripFeedback, aspect feedback.Aspect) []domain.Signal {
	if aspect == feedback.LongTransfers {
		signal := domain.NewSignal(submission.UserID, domain.FeedbackPositive,
			domain.TravelBurden, "long_transfers", submission.SubmittedAt)
		signal.TripID = submission.TripID
		signal.SourceEvent = "feedback:" + submission.TripID + ":too_much_travel"
		return []domain.Signal{signal}
	}
	return aspectSignal(submission, aspect, domain.FeedbackNegative)
}

func paceSignals(submission feedback.TripFeedback) []domain.Signal {
	var value string
	switch submission.Pace {
	case feedback.TooBusy:
		value = "relaxed"
	case feedback.TooSlow:
		value = "packed"
	default:
		return nil
	}
	signal := domain.NewSignal(submission.UserID, domain.FeedbackPositive, domain.TripPace, value, submission.SubmittedAt)
	signal.TripID = submission.TripID
	signal.SourceEvent = "feedback:" + submission.TripID + ":pace"
	return []domain.Signal{signal}
}

func burdenSignals(submission feedback.TripFeedback) []domain.Signal {
	if submission.TravelBurden != feedback.TooMuchTravel {
		return nil
	}
	signal := domain.NewSignal(submission.UserID, domain.FeedbackPositive,
		domain.TravelBurden, "too_much", submission.SubmittedAt)
	signal.TripID = submission.TripID
	signal.SourceEvent = "feedback:" + submission.TripID + ":burden"
	return []domain.Signal{signal}
}

func SuppressionUntil(now time.Time, thresholds Thresholds) time.Time {
	return now.UTC().Add(thresholds.SuppressionWindow)
}
