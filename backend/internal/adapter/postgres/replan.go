package postgres

import (
	"context"
	"encoding/json"
	"time"

	queries "github.com/SemiuAdesina/nuvro/backend/internal/adapter/postgres/sqlc"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/money"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/replan"
)

func (s *Store) SaveProposal(ctx context.Context, proposal replan.Proposal) error {
	if err := proposal.Validate(); err != nil {
		return err
	}
	proposalID, err := identifier(proposal.ID)
	if err != nil {
		return err
	}
	tripID, err := identifier(proposal.TripID)
	if err != nil {
		return err
	}
	changes, err := json.Marshal(proposal.Changes)
	if err != nil {
		return err
	}
	params := queries.InsertReplanProposalParams{
		ID: proposalID, TripID: tripID, Trigger: string(proposal.Trigger),
		ItineraryVersion: bounded(proposal.ItineraryVersion), Changes: changes,
		Reasons: reasonCodes(proposal.Reasons), Preserved: constraintCodes(proposal.Preserved),
		Summary: proposal.Summary, ExpiresAt: proposal.ExpiresAt.UTC(),
	}
	if proposal.HasBudgetDelta {
		minor := proposal.BudgetDelta.Minor()
		params.BudgetDeltaMinor = &minor
		params.BudgetCurrency = proposal.BudgetDelta.Currency().String()
	}
	return databaseError(s.queries(ctx).InsertReplanProposal(ctx, params))
}

func (s *Store) Proposal(ctx context.Context, proposalID string) (replan.Proposal, error) {
	identity, err := identifier(proposalID)
	if err != nil {
		return replan.Proposal{}, err
	}
	row, err := s.queries(ctx).GetReplanProposal(ctx, identity)
	if err != nil {
		return replan.Proposal{}, databaseError(err)
	}
	var changes []replan.Change
	if err := json.Unmarshal(row.Changes, &changes); err != nil {
		return replan.Proposal{}, err
	}
	proposal := replan.Proposal{
		ID: uuidText(row.ID), TripID: uuidText(row.TripID), Trigger: replan.Trigger(row.Trigger),
		ItineraryVersion: int(row.ItineraryVersion), Changes: changes,
		Reasons: toReasons(row.Reasons), Preserved: toConstraints(row.Preserved),
		Summary: row.Summary, CreatedAt: row.CreatedAt.UTC(), ExpiresAt: row.ExpiresAt.UTC(),
	}
	if row.AppliedAt != nil {
		applied := row.AppliedAt.UTC()
		proposal.AppliedAt = &applied
	}
	if row.BudgetDeltaMinor != nil && row.BudgetCurrency != "" {
		amount, amountErr := money.New(*row.BudgetDeltaMinor, money.Currency(row.BudgetCurrency))
		if amountErr == nil {
			proposal.BudgetDelta = amount
			proposal.HasBudgetDelta = true
		}
	}
	return proposal, nil
}

func (s *Store) MarkProposalApplied(ctx context.Context, proposalID, userID string, at time.Time) error {
	identity, err := identifier(proposalID)
	if err != nil {
		return err
	}
	applier, err := identifier(userID)
	if err != nil {
		return err
	}
	applied := at.UTC()
	return databaseError(s.queries(ctx).MarkReplanProposalApplied(ctx, queries.MarkReplanProposalAppliedParams{
		ID: identity, AppliedAt: &applied, AppliedBy: applier,
	}))
}

func reasonCodes(reasons []replan.Reason) []string {
	codes := make([]string, 0, len(reasons))
	for _, reason := range reasons {
		codes = append(codes, string(reason))
	}
	return codes
}

func constraintCodes(constraints []replan.Constraint) []string {
	codes := make([]string, 0, len(constraints))
	for _, constraint := range constraints {
		codes = append(codes, string(constraint))
	}
	return codes
}

func toReasons(codes []string) []replan.Reason {
	reasons := make([]replan.Reason, 0, len(codes))
	for _, code := range codes {
		reasons = append(reasons, replan.Reason(code))
	}
	return reasons
}

func toConstraints(codes []string) []replan.Constraint {
	constraints := make([]replan.Constraint, 0, len(codes))
	for _, code := range codes {
		constraints = append(constraints, replan.Constraint(code))
	}
	return constraints
}
