package postgres

import (
	"context"

	queries "github.com/SemiuAdesina/nuvro/backend/internal/adapter/postgres/sqlc"
	domain "github.com/SemiuAdesina/nuvro/backend/internal/domain/commerce"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/money"
)

func (s *Store) SaveConversion(ctx context.Context, conversion domain.Conversion) (domain.Conversion, error) {
	if err := conversion.Validate(); err != nil {
		return domain.Conversion{}, err
	}
	conversionID, err := identifier(conversion.ID)
	if err != nil {
		return domain.Conversion{}, err
	}
	params := queries.UpsertConversionParams{
		ID: conversionID, Partner: conversion.Partner, ProviderConversionID: conversion.ProviderConversionID,
		ProductType: string(conversion.Product), Status: string(conversion.Status),
		BookingReference: conversion.BookingReference, OccurredAt: conversion.OccurredAt.UTC(),
	}
	if conversion.HandoffID != "" {
		handoffID, handoffErr := identifier(conversion.HandoffID)
		if handoffErr != nil {
			return domain.Conversion{}, handoffErr
		}
		params.HandoffID = handoffID
	}
	if conversion.HasGrossValue {
		gross := conversion.GrossValue.Minor()
		params.GrossValueMinor = &gross
		params.GrossValueCurrency = conversion.GrossValue.Currency().String()
	}
	if conversion.HasCommission {
		commission := conversion.Commission.Minor()
		params.CommissionMinor = &commission
		params.CommissionCurrency = conversion.Commission.Currency().String()
	}
	row, err := s.queries(ctx).UpsertConversion(ctx, params)
	if err != nil {
		return domain.Conversion{}, databaseError(err)
	}
	return toConversion(queries.GetConversionRow(row))
}

func (s *Store) Conversion(ctx context.Context, partner, providerConversionID string) (domain.Conversion, error) {
	row, err := s.queries(ctx).GetConversion(ctx, queries.GetConversionParams{
		Partner: partner, ProviderConversionID: providerConversionID,
	})
	if err != nil {
		return domain.Conversion{}, databaseError(err)
	}
	return toConversion(row)
}

func toConversion(row queries.GetConversionRow) (domain.Conversion, error) {
	conversion := domain.Conversion{
		ID: uuidText(row.ID), Partner: row.Partner, ProviderConversionID: row.ProviderConversionID,
		Product: domain.ProductType(row.ProductType), Status: domain.ConversionStatus(row.Status),
		BookingReference: row.BookingReference, OccurredAt: row.OccurredAt.UTC(), UpdatedAt: row.UpdatedAt.UTC(),
	}
	if row.HandoffID.Valid {
		conversion.HandoffID = uuidText(row.HandoffID)
	}
	if row.GrossValueMinor != nil && row.GrossValueCurrency != "" {
		amount, err := money.New(*row.GrossValueMinor, money.Currency(row.GrossValueCurrency))
		if err != nil {
			return domain.Conversion{}, err
		}
		conversion.GrossValue = amount
		conversion.HasGrossValue = true
	}
	if row.CommissionMinor != nil && row.CommissionCurrency != "" {
		amount, err := money.New(*row.CommissionMinor, money.Currency(row.CommissionCurrency))
		if err != nil {
			return domain.Conversion{}, err
		}
		conversion.Commission = amount
		conversion.HasCommission = true
	}
	return conversion, nil
}
