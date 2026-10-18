package postgres

import (
	"context"
	"time"

	queries "github.com/SemiuAdesina/nuvro/backend/internal/adapter/postgres/sqlc"
	"github.com/SemiuAdesina/nuvro/backend/internal/apperror"
	domain "github.com/SemiuAdesina/nuvro/backend/internal/domain/commerce"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/money"
)

func (s *Store) SaveHandoff(ctx context.Context, handoff domain.Handoff) error {
	if err := handoff.Validate(); err != nil {
		return err
	}
	handoffID, err := identifier(handoff.ID)
	if err != nil {
		return err
	}
	userID, err := identifier(handoff.UserID)
	if err != nil {
		return err
	}
	params := queries.InsertHandoffParams{
		ID: handoffID, OfferID: handoff.OfferID, UserID: userID, Partner: handoff.Partner,
		ProductType: string(handoff.Product), DestinationUrl: handoff.DestinationURL,
		Title: handoff.Title, PriceMinor: handoff.Price.Minor(),
		PriceCurrency: handoff.Price.Currency().String(), AttributionToken: handoff.AttributionToken,
		ExpiresAt: handoff.ExpiresAt.UTC(),
	}
	if handoff.TripID != "" {
		tripID, tripErr := identifier(handoff.TripID)
		if tripErr != nil {
			return tripErr
		}
		params.TripID = tripID
	}
	return databaseError(s.queries(ctx).InsertHandoff(ctx, params))
}

func (s *Store) Handoff(ctx context.Context, handoffID string) (domain.Handoff, error) {
	identity, err := identifier(handoffID)
	if err != nil {
		return domain.Handoff{}, err
	}
	row, err := s.queries(ctx).GetHandoff(ctx, identity)
	if err != nil {
		return domain.Handoff{}, databaseError(err)
	}
	price, err := money.New(row.PriceMinor, money.Currency(row.PriceCurrency))
	if err != nil {
		return domain.Handoff{}, err
	}
	handoff := domain.Handoff{
		ID: uuidText(row.ID), OfferID: row.OfferID, UserID: uuidText(row.UserID),
		Partner: row.Partner, Product: domain.ProductType(row.ProductType),
		DestinationURL: row.DestinationUrl, Title: row.Title, Price: price,
		Disclosure:       domain.Disclosure{Code: domain.EligibleBookingDisclosure, Partner: row.Partner},
		AttributionToken: row.AttributionToken, CreatedAt: row.CreatedAt.UTC(), ExpiresAt: row.ExpiresAt.UTC(),
	}
	if row.TripID.Valid {
		handoff.TripID = uuidText(row.TripID)
	}
	return handoff, nil
}

func (s *Store) RecordClick(ctx context.Context, click domain.Click) error {
	if err := click.Validate(); err != nil {
		return err
	}
	clickID, err := identifier(click.ID)
	if err != nil {
		return err
	}
	handoffID, err := identifier(click.HandoffID)
	if err != nil {
		return err
	}
	return databaseError(s.queries(ctx).RecordClick(ctx, queries.RecordClickParams{
		ID: clickID, HandoffID: handoffID, Partner: click.Partner,
		ProductType: string(click.Product), OccurredAt: click.OccurredAt.UTC(),
	}))
}

func (s *Store) ClickCount(ctx context.Context, handoffID string) (int64, error) {
	identity, err := identifier(handoffID)
	if err != nil {
		return 0, err
	}
	count, err := s.queries(ctx).ClickCount(ctx, identity)
	return count, databaseError(err)
}

func (s *Store) ClaimWebhookEvent(ctx context.Context, eventID, partner, providerEventID, eventType, status string) (bool, error) {
	identity, err := identifier(eventID)
	if err != nil {
		return false, err
	}
	_, err = s.queries(ctx).ClaimWebhookEvent(ctx, queries.ClaimWebhookEventParams{
		ID: identity, Partner: partner, ProviderEventID: providerEventID,
		EventType: eventType, NormalizedStatus: status,
	})
	if err != nil {
		if apperror.CodeOf(databaseError(err)) == apperror.CodeNotFound {
			return false, nil
		}
		return false, databaseError(err)
	}
	return true, nil
}

func (s *Store) MarkWebhookProcessed(ctx context.Context, partner, providerEventID string, at time.Time) error {
	processed := at.UTC()
	return databaseError(s.queries(ctx).MarkWebhookProcessed(ctx, queries.MarkWebhookProcessedParams{
		Partner: partner, ProviderEventID: providerEventID, ProcessedAt: &processed,
	}))
}
