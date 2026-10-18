package travel

import (
	"context"
	"errors"

	"github.com/SemiuAdesina/nuvro/backend/internal/apperror"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/geo"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/itinerary"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/replan"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
	"github.com/SemiuAdesina/nuvro/backend/internal/engine/replanning"
)

type ReplanRequest struct {
	TripID   string
	Trigger  string
	Message  string
	Location *geo.Point
	Language string
}

type ReplanResult struct {
	Proposal replan.Proposal
	Original []itinerary.Item
}

const replanBufferMinutes = 15

func (s *Service) Replan(ctx context.Context, principal user.Principal, request ReplanRequest) (ReplanResult, error) {
	if s.planner == nil {
		return ReplanResult{}, apperror.New(apperror.CodeUnavailable, "Replanning evidence is unavailable.")
	}
	record, place, err := s.context(ctx, principal, request.TripID)
	if err != nil {
		return ReplanResult{}, err
	}
	if err := s.assertEditor(ctx, principal, request.TripID); err != nil {
		return ReplanResult{}, err
	}
	plan, hasPlan := s.plan(ctx, request.TripID)
	if !hasPlan || len(plan.Items) == 0 {
		return ReplanResult{}, apperror.New(apperror.CodeInvalidState, "This trip has no itinerary to adjust.")
	}
	trigger := replan.Trigger(request.Trigger)
	if !trigger.Valid() {
		return ReplanResult{}, apperror.New(apperror.CodeInvalidInput, "That replanning trigger is not recognised.")
	}
	window, _ := s.forecast(ctx, place, record, request.Language)
	proposal := s.planner.Propose(ctx, replanning.Request{
		TripID: request.TripID, Trigger: trigger, Itinerary: plan, Weather: window,
		Now: s.timer.Now(), CurrentLocation: request.Location, Message: request.Message,
		Language: request.Language, Currency: record.Currency, Timezone: place.Timezone,
		BudgetCeiling: nil, ReducedWalking: s.reducedWalking(ctx, principal), Pace: s.pace(ctx, principal),
	})
	if len(proposal.Changes) == 0 {
		return ReplanResult{Proposal: proposal, Original: plan.Items}, nil
	}
	if err := s.persist(ctx, &proposal); err != nil {
		return ReplanResult{}, err
	}
	return ReplanResult{Proposal: proposal, Original: plan.Items}, nil
}

func (s *Service) persist(ctx context.Context, proposal *replan.Proposal) error {
	identifier, err := s.identifiers.Generate()
	if err != nil {
		return apperror.Wrap(apperror.CodeInternal, "", err)
	}
	proposal.ID = identifier
	if err := s.proposals.SaveProposal(ctx, *proposal); err != nil {
		return err
	}
	return nil
}

func (s *Service) ApplyReplan(ctx context.Context, principal user.Principal, tripID, proposalID string) (itinerary.Itinerary, error) {
	if _, err := s.trips.Read(ctx, principal, tripID); err != nil {
		return itinerary.Itinerary{}, err
	}
	if err := s.assertEditor(ctx, principal, tripID); err != nil {
		return itinerary.Itinerary{}, err
	}
	proposal, plan, err := s.applicable(ctx, tripID, proposalID)
	if err != nil {
		return itinerary.Itinerary{}, err
	}
	updated := replanning.Apply(plan, proposal)
	violations := itinerary.Validate(updated, itinerary.Constraints{
		BufferMinutes: replanBufferMinutes, Locked: lockedByID(plan),
	})
	if len(violations) > 0 {
		return itinerary.Itinerary{}, apperror.Wrap(apperror.CodeInvalidState,
			"That change no longer produces a workable plan.", violationError(violations))
	}
	saved, err := s.itineraries.ReplaceItineraryItems(ctx, tripID, plan.Version, updated.Items)
	if err != nil {
		return itinerary.Itinerary{}, err
	}
	if err := s.proposals.MarkProposalApplied(ctx, proposalID, principal.UserID, s.timer.Now()); err != nil {
		return itinerary.Itinerary{}, err
	}
	return saved, nil
}

func (s *Service) applicable(ctx context.Context, tripID, proposalID string) (replan.Proposal, itinerary.Itinerary, error) {
	proposal, err := s.proposals.Proposal(ctx, proposalID)
	if err != nil || proposal.TripID != tripID {
		return replan.Proposal{}, itinerary.Itinerary{}, apperror.New(apperror.CodeNotFound, "Proposal not found.")
	}
	if proposal.Applied() {
		return replan.Proposal{}, itinerary.Itinerary{},
			apperror.New(apperror.CodeConflict, "This proposal has already been applied.")
	}
	if proposal.Stale(s.timer.Now()) {
		return replan.Proposal{}, itinerary.Itinerary{},
			apperror.New(apperror.CodeInvalidState, "This suggestion has expired. Ask for a new one.")
	}
	plan, hasPlan := s.plan(ctx, tripID)
	if !hasPlan {
		return replan.Proposal{}, itinerary.Itinerary{},
			apperror.New(apperror.CodeNotFound, "No itinerary has been generated for this trip.")
	}
	if plan.Version != proposal.ItineraryVersion {
		return replan.Proposal{}, itinerary.Itinerary{},
			apperror.New(apperror.CodeConflict, "The plan changed since this suggestion was made.")
	}
	if lockErr := assertLocksPreserved(plan, proposal); lockErr != nil {
		return replan.Proposal{}, itinerary.Itinerary{}, lockErr
	}
	return proposal, plan, nil
}

func lockedByID(plan itinerary.Itinerary) map[string]itinerary.Item {
	locked := make(map[string]itinerary.Item)
	for _, item := range plan.LockedItems() {
		locked[item.ID] = item
	}
	return locked
}

func violationError(violations []itinerary.Violation) error {
	return errors.New("replan: " + string(violations[0].Code))
}

func assertLocksPreserved(plan itinerary.Itinerary, proposal replan.Proposal) error {
	locked := make(map[string]itinerary.Item, len(plan.Items))
	for _, item := range plan.LockedItems() {
		locked[item.ID] = item
	}
	for _, change := range proposal.Changes {
		if change.Original == nil {
			continue
		}
		original, found := locked[change.Original.ID]
		if !found {
			continue
		}
		if change.Proposed == nil || !change.Proposed.StartAt.Equal(original.StartAt) {
			return apperror.New(apperror.CodeInvalidState, "That change would move a reserved item.")
		}
	}
	return nil
}

func (s *Service) assertEditor(ctx context.Context, principal user.Principal, tripID string) error {
	role, err := s.trips.Role(ctx, tripID, principal.UserID)
	if err != nil {
		return err
	}
	if !role.CanEdit() {
		return apperror.New(apperror.CodeForbidden, "Editing this trip is not permitted.")
	}
	return nil
}

func (s *Service) reducedWalking(ctx context.Context, principal user.Principal) bool {
	if s.preferences == nil {
		return false
	}
	profile, err := s.preferences.Profile(ctx, principal.UserID)
	if err != nil {
		return false
	}
	for _, preference := range profile.AccessibilityPreferences {
		if preference == "mobility" || preference == "reduced_walking" {
			return true
		}
	}
	return false
}

func (s *Service) pace(ctx context.Context, principal user.Principal) string {
	if s.preferences == nil {
		return ""
	}
	preferences, err := s.preferences.Preferences(ctx, principal.UserID)
	if err != nil {
		return ""
	}
	return preferences.Pace
}
