package feasibility

import (
	"context"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/destination"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/discovery"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/feasibility"
)

func (e *Engine) access(ctx context.Context, intent discovery.Intent, target destination.Destination) feasibility.Access {
	if e.entry == nil {
		return feasibility.UnknownAccessAssessment("entry_provider_unavailable")
	}
	if !intent.Traveler.HasPassport() {
		return feasibility.UnknownAccessAssessment("passport_context_required")
	}
	best := feasibility.UnknownAccessAssessment("entry_evidence_unavailable")
	for _, passport := range intent.Traveler.PassportCountries {
		candidate := e.accessForPassport(ctx, intent, target, passport)
		if candidate.Rank() > best.Rank() {
			best = candidate
		}
	}
	return best
}

func (e *Engine) accessForPassport(ctx context.Context, intent discovery.Intent, target destination.Destination, passport string) feasibility.Access {
	evidence, err := e.entry.Entry(ctx, EntryQuery{
		PassportCountry: passport, ResidenceCountry: intent.Traveler.ResidenceCountry,
		ResidencePermits: intent.Traveler.ResidencePermits, OriginCountry: intent.Traveler.ResidenceCountry,
		DestinationCountry: target.CountryCode, Start: intent.Window.Start, End: intent.Window.End,
	})
	if err != nil {
		return feasibility.UnknownAccessAssessment("entry_evidence_unavailable")
	}
	assessment := feasibility.Access{
		Status: statusFor(evidence), EntryMode: evidence.EntryMode, Preparation: evidence.Preparation,
		TransitRisk: transitOrUnknown(evidence.TransitRisk), Confidence: confidenceOrUnknown(evidence.Confidence),
		Reasons: evidence.Reasons, SourceLabel: evidence.SourceLabel, ObservedAt: evidence.ObservedAt,
		PassportUsed: passport,
	}
	return assessment
}

func statusFor(evidence EntryEvidence) feasibility.AccessStatus {
	if evidence.EntryMode == feasibility.UnknownEntry || evidence.Confidence == feasibility.UnknownConfidence {
		return feasibility.UnknownAccess
	}
	switch evidence.Preparation {
	case feasibility.NoPreparation:
		return feasibility.Easy
	case feasibility.LowPreparation:
		return preparationByMode(evidence.EntryMode, feasibility.Easy, feasibility.Manageable)
	case feasibility.MediumPreparation:
		return feasibility.Manageable
	case feasibility.HighPreparation:
		return feasibility.PreparationRequired
	default:
		return feasibility.UnknownAccess
	}
}

func preparationByMode(mode feasibility.EntryMode, simple, involved feasibility.AccessStatus) feasibility.AccessStatus {
	switch mode {
	case feasibility.VisaFree, feasibility.ETA:
		return simple
	default:
		return involved
	}
}

func transitOrUnknown(risk feasibility.TransitRisk) feasibility.TransitRisk {
	switch risk {
	case feasibility.NoTransitRisk, feasibility.TransitReviewNeeded:
		return risk
	default:
		return feasibility.UnknownTransit
	}
}

func confidenceOrUnknown(confidence feasibility.Confidence) feasibility.Confidence {
	if confidence.Valid() {
		return confidence
	}
	return feasibility.UnknownConfidence
}
