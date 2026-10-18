package postgres

import (
	"context"
	"time"

	queries "github.com/SemiuAdesina/nuvro/backend/internal/adapter/postgres/sqlc"
	domain "github.com/SemiuAdesina/nuvro/backend/internal/domain/personalization"
)

const signalPageSize = 500

func (s *Store) RecordSignal(ctx context.Context, signalID string, signal domain.Signal) error {
	if err := signal.Validate(); err != nil {
		return err
	}
	identity, err := identifier(signalID)
	if err != nil {
		return err
	}
	userID, err := identifier(signal.UserID)
	if err != nil {
		return err
	}
	params := queries.RecordSignalParams{
		ID: identity, UserID: userID, SignalType: string(signal.Type),
		SubjectType: string(signal.Subject), SubjectID: signal.SubjectID,
		Strength: string(signal.Strength), SourceEvent: signal.SourceEvent,
		OccurredAt: signal.OccurredAt.UTC(),
	}
	if signal.TripID != "" {
		tripID, tripErr := identifier(signal.TripID)
		if tripErr != nil {
			return tripErr
		}
		params.TripID = tripID
	}
	return databaseError(s.queries(ctx).RecordSignal(ctx, params))
}

func (s *Store) Signals(ctx context.Context, userID string, since time.Time) ([]domain.Signal, error) {
	identity, err := identifier(userID)
	if err != nil {
		return nil, err
	}
	rows, err := s.queries(ctx).ListSignals(ctx, queries.ListSignalsParams{
		UserID: identity, OccurredAt: since.UTC(), Limit: signalPageSize,
	})
	if err != nil {
		return nil, databaseError(err)
	}
	signals := make([]domain.Signal, 0, len(rows))
	for _, row := range rows {
		signal := domain.Signal{
			ID: uuidText(row.ID), UserID: uuidText(row.UserID), Type: domain.SignalType(row.SignalType),
			Subject: domain.SubjectType(row.SubjectType), SubjectID: row.SubjectID,
			Strength: domain.Strength(row.Strength), SourceEvent: row.SourceEvent,
			OccurredAt: row.OccurredAt.UTC(),
		}
		if row.TripID.Valid {
			signal.TripID = uuidText(row.TripID)
		}
		signals = append(signals, signal)
	}
	return signals, nil
}

func (s *Store) ClearSignals(ctx context.Context, userID string) error {
	identity, err := identifier(userID)
	if err != nil {
		return err
	}
	return databaseError(s.queries(ctx).DeleteSignals(ctx, identity))
}

func (s *Store) SaveLearnedPreference(ctx context.Context, preferenceID string, preference domain.LearnedPreference) error {
	if err := preference.Validate(); err != nil {
		return err
	}
	identity, err := identifier(preferenceID)
	if err != nil {
		return err
	}
	userID, err := identifier(preference.UserID)
	if err != nil {
		return err
	}
	return databaseError(s.queries(ctx).UpsertLearnedPreference(ctx, queries.UpsertLearnedPreferenceParams{
		ID: identity, UserID: userID, PreferenceKey: string(preference.Key),
		Confidence: string(preference.Confidence), Score: bounded(preference.Score),
		EvidenceCount: bounded(preference.EvidenceCount), ExplanationCode: string(preference.Explanation),
		Status: string(preference.Status), RulesVersion: preference.RulesVersion,
		FirstObservedAt: preference.FirstObservedAt.UTC(), LastObservedAt: preference.LastObservedAt.UTC(),
	}))
}

func (s *Store) LearnedPreferences(ctx context.Context, userID string) ([]domain.LearnedPreference, error) {
	identity, err := identifier(userID)
	if err != nil {
		return nil, err
	}
	rows, err := s.queries(ctx).ListLearnedPreferences(ctx, identity)
	if err != nil {
		return nil, databaseError(err)
	}
	preferences := make([]domain.LearnedPreference, 0, len(rows))
	for _, row := range rows {
		preferences = append(preferences, toLearnedPreference(row))
	}
	return preferences, nil
}

func (s *Store) LearnedPreference(ctx context.Context, preferenceID string) (domain.LearnedPreference, error) {
	identity, err := identifier(preferenceID)
	if err != nil {
		return domain.LearnedPreference{}, err
	}
	row, err := s.queries(ctx).GetLearnedPreference(ctx, identity)
	if err != nil {
		return domain.LearnedPreference{}, databaseError(err)
	}
	return toLearnedPreference(row), nil
}

func (s *Store) UpdateLearnedPreferenceStatus(ctx context.Context, preferenceID string, status domain.Status) error {
	identity, err := identifier(preferenceID)
	if err != nil {
		return err
	}
	return databaseError(s.queries(ctx).UpdateLearnedPreferenceStatus(ctx, queries.UpdateLearnedPreferenceStatusParams{
		ID: identity, Status: string(status),
	}))
}

func (s *Store) ClearLearnedPreferences(ctx context.Context, userID string) error {
	identity, err := identifier(userID)
	if err != nil {
		return err
	}
	return databaseError(s.queries(ctx).DeleteLearnedPreferences(ctx, identity))
}

func (s *Store) SetHistoryLearning(ctx context.Context, userID string, enabled bool) error {
	identity, err := identifier(userID)
	if err != nil {
		return err
	}
	return databaseError(s.queries(ctx).SetHistoryLearning(ctx, queries.SetHistoryLearningParams{
		UserID: identity, HistoryLearningEnabled: enabled,
	}))
}

func (s *Store) HistoryLearning(ctx context.Context, userID string) (bool, error) {
	identity, err := identifier(userID)
	if err != nil {
		return false, err
	}
	enabled, err := s.queries(ctx).GetHistoryLearning(ctx, identity)
	if err != nil {
		return false, databaseError(err)
	}
	return enabled, nil
}

func toLearnedPreference(row queries.LearnedPreference) domain.LearnedPreference {
	return domain.LearnedPreference{
		ID: uuidText(row.ID), UserID: uuidText(row.UserID), Key: domain.Key(row.PreferenceKey),
		Confidence: domain.Confidence(row.Confidence), Score: int(row.Score),
		EvidenceCount: int(row.EvidenceCount), Explanation: domain.ExplanationCode(row.ExplanationCode),
		Status: domain.Status(row.Status), RulesVersion: row.RulesVersion,
		FirstObservedAt: row.FirstObservedAt.UTC(), LastObservedAt: row.LastObservedAt.UTC(),
		UpdatedAt: row.UpdatedAt.UTC(),
	}
}
