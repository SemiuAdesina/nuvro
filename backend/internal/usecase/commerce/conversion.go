package commerce

import (
	"context"
	"errors"

	"github.com/SemiuAdesina/nuvro/backend/internal/apperror"
	domain "github.com/SemiuAdesina/nuvro/backend/internal/domain/commerce"
	engine "github.com/SemiuAdesina/nuvro/backend/internal/engine/commerce"
)

type ConversionResult struct {
	Conversion domain.Conversion
	Duplicate  bool
}

func (s *Service) ProcessWebhook(ctx context.Context, event domain.WebhookEvent) (ConversionResult, error) {
	if err := event.Validate(); err != nil {
		return ConversionResult{}, apperror.Wrap(apperror.CodeInvalidInput, "The event payload is not valid.", err)
	}
	eventID, err := s.identifiers.Generate()
	if err != nil {
		return ConversionResult{}, apperror.Wrap(apperror.CodeInternal, "", err)
	}
	claimed, err := s.conversions.ClaimWebhookEvent(ctx, eventID, event.Partner, event.EventID,
		event.EventType, string(event.Conversion.Status))
	if err != nil {
		return ConversionResult{}, err
	}
	if !claimed {
		return s.alreadyProcessed(ctx, event), nil
	}
	conversion, err := s.reconcile(ctx, event)
	if err != nil {
		return ConversionResult{}, err
	}
	if err := s.conversions.MarkWebhookProcessed(ctx, event.Partner, event.EventID, s.timer.Now()); err != nil {
		return ConversionResult{}, err
	}
	return ConversionResult{Conversion: conversion}, nil
}

func (s *Service) alreadyProcessed(ctx context.Context, event domain.WebhookEvent) ConversionResult {
	existing, err := s.conversions.Conversion(ctx, event.Partner, event.Conversion.ProviderConversionID)
	if err != nil {
		return ConversionResult{Duplicate: true}
	}
	return ConversionResult{Conversion: existing, Duplicate: true}
}

func (s *Service) reconcile(ctx context.Context, event domain.WebhookEvent) (domain.Conversion, error) {
	incoming := event.Conversion
	existing, err := s.conversions.Conversion(ctx, event.Partner, incoming.ProviderConversionID)
	switch {
	case err == nil:
		if !engine.Supersedes(existing.Status, incoming.Status) {
			return existing, nil
		}
		incoming.ID = existing.ID
	case apperror.CodeOf(err) == apperror.CodeNotFound:
		identifier, generateErr := s.identifiers.Generate()
		if generateErr != nil {
			return domain.Conversion{}, apperror.Wrap(apperror.CodeInternal, "", generateErr)
		}
		incoming.ID = identifier
	default:
		return domain.Conversion{}, err
	}
	return s.conversions.SaveConversion(ctx, incoming)
}

func NormalizeStatus(providerCode string) (domain.ConversionStatus, error) {
	status, err := engine.NormalizeStatus(providerCode)
	if err != nil {
		if errors.Is(err, engine.ErrUnknownEvent) {
			return "", apperror.Wrap(apperror.CodeInvalidInput, "That event type is not recognised.", err)
		}
		return "", err
	}
	return status, nil
}
