package commerce_test

import (
	"testing"

	"github.com/SemiuAdesina/nuvro/backend/internal/apperror"
	domain "github.com/SemiuAdesina/nuvro/backend/internal/domain/commerce"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/pricewatch"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
	usecase "github.com/SemiuAdesina/nuvro/backend/internal/usecase/commerce"
)

func TestWebhookProcessingIsIdempotent(t *testing.T) {
	conversions := newConversionStore()
	service := newService(t, usecase.Dependencies{Conversions: conversions})
	event := domain.WebhookEvent{
		Partner: "fixture", EventID: "evt-1", EventType: "booking.completed", ReceivedAt: now,
		Conversion: domain.Conversion{
			Partner: "fixture", ProviderConversionID: "conv-1", Product: domain.Flight,
			Status: domain.BookingCompleted, BookingReference: "ABC123XYZ",
			GrossValue: amount(t, 50_000), HasGrossValue: true,
			Commission: amount(t, 2_500), HasCommission: true, OccurredAt: now,
		},
	}
	first, err := service.ProcessWebhook(t.Context(), event)
	if err != nil {
		t.Fatal(err)
	}
	if first.Duplicate || first.Conversion.Status != domain.BookingCompleted {
		t.Fatalf("first delivery = %+v", first)
	}
	second, err := service.ProcessWebhook(t.Context(), event)
	if err != nil {
		t.Fatal(err)
	}
	if !second.Duplicate {
		t.Fatal("a repeated delivery created a second conversion")
	}
	if len(conversions.conversions) != 1 {
		t.Fatalf("stored %d conversions", len(conversions.conversions))
	}
	if conversions.processed != 1 {
		t.Fatalf("processed %d times", conversions.processed)
	}
}

func TestLateDuplicatesDoNotReopenACancelledBooking(t *testing.T) {
	conversions := newConversionStore()
	service := newService(t, usecase.Dependencies{Conversions: conversions})
	base := domain.Conversion{
		Partner: "fixture", ProviderConversionID: "conv-1", Product: domain.Stay,
		Status: domain.BookingCancelled, OccurredAt: now,
	}
	if _, err := service.ProcessWebhook(t.Context(), domain.WebhookEvent{
		Partner: "fixture", EventID: "evt-cancel", ReceivedAt: now, Conversion: base,
	}); err != nil {
		t.Fatal(err)
	}
	late := base
	late.Status = domain.BookingCreated
	if _, err := service.ProcessWebhook(t.Context(), domain.WebhookEvent{
		Partner: "fixture", EventID: "evt-create", ReceivedAt: now, Conversion: late,
	}); err != nil {
		t.Fatal(err)
	}
	stored := conversions.conversions["fixture|conv-1"]
	if stored.Status != domain.BookingCancelled {
		t.Fatalf("a late creation event reopened the booking: %q", stored.Status)
	}
}

func TestMalformedEventsAreRefused(t *testing.T) {
	service := newService(t, usecase.Dependencies{Conversions: newConversionStore()})
	if _, err := service.ProcessWebhook(t.Context(), domain.WebhookEvent{}); apperror.CodeOf(err) != apperror.CodeInvalidInput {
		t.Fatal("an empty event was accepted")
	}
	if _, err := usecase.NormalizeStatus("something_new"); apperror.CodeOf(err) != apperror.CodeInvalidInput {
		t.Fatal("an unknown provider status was accepted")
	}
	if status, err := usecase.NormalizeStatus("confirmed"); err != nil || status != domain.BookingCompleted {
		t.Fatalf("%q %v", status, err)
	}
}

func TestPriceMonitoringRequiresEntitlement(t *testing.T) {
	watches := newWatchStore()
	blocked := newService(t, usecase.Dependencies{Watches: watches, Entitlements: entitlements{}})
	request := usecase.WatchRequest{
		Product: string(domain.Flight),
		Target: pricewatch.Target{
			Origin: "LIS", Destination: "FNC", StartDate: now.AddDate(0, 1, 0),
			EndDate: now.AddDate(0, 1, 6), Adults: 2,
		},
		BaselineMinor: 20_000, Currency: "EUR",
	}
	if _, err := blocked.CreateWatch(t.Context(), principal(), request); apperror.CodeOf(err) != apperror.CodeForbidden {
		t.Fatalf("error = %v", err)
	}
	unconfigured := newService(t, usecase.Dependencies{Watches: watches})
	if _, err := unconfigured.CreateWatch(t.Context(), principal(), request); apperror.CodeOf(err) != apperror.CodeUnavailable {
		t.Fatal("an unconfigured entitlement backend allowed a watch")
	}
	allowed := newService(t, usecase.Dependencies{Watches: watches, Entitlements: entitlements{allowed: true}})
	watch, err := allowed.CreateWatch(t.Context(), principal(), request)
	if err != nil {
		t.Fatal(err)
	}
	if watch.Status != pricewatch.Active || watch.Threshold.PercentageDrop == 0 {
		t.Fatalf("watch = %+v", watch)
	}
	if !watch.NextCheckAt.After(now) || !watch.ExpiresAt.After(now) {
		t.Fatalf("watch scheduling = %+v", watch)
	}
}

func TestWatchLimitAndOwnership(t *testing.T) {
	watches := newWatchStore()
	service := newService(t, usecase.Dependencies{Watches: watches, Entitlements: entitlements{allowed: true}})
	request := usecase.WatchRequest{
		Product: string(domain.Flight),
		Target: pricewatch.Target{
			Origin: "LIS", Destination: "FNC", StartDate: now.AddDate(0, 1, 0),
			EndDate: now.AddDate(0, 1, 6), Adults: 2,
		},
		BaselineMinor: 20_000, Currency: "EUR",
	}
	for range pricewatch.MaximumPerUser {
		if _, err := service.CreateWatch(t.Context(), principal(), request); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := service.CreateWatch(t.Context(), principal(), request); apperror.CodeOf(err) != apperror.CodeConflict {
		t.Fatal("the watch limit was not enforced")
	}
	stored, err := service.Watches(t.Context(), principal())
	if err != nil || len(stored) != pricewatch.MaximumPerUser {
		t.Fatalf("%d watches %v", len(stored), err)
	}
	stranger := user.Principal{UserID: "0c2f21b0-4a55-4b39-9f0b-2c7f9c1a1a99"}
	if err := service.DeleteWatch(t.Context(), stranger, stored[0].ID); apperror.CodeOf(err) != apperror.CodeNotFound {
		t.Fatal("another traveller deleted a watch")
	}
	if err := service.DeleteWatch(t.Context(), principal(), stored[0].ID); err != nil {
		t.Fatal(err)
	}
}

func TestWatchSchedulingTightensNearDeparture(t *testing.T) {
	far := usecase.NextCheck(now, now.AddDate(0, 3, 0))
	medium := usecase.NextCheck(now, now.AddDate(0, 0, 20))
	near := usecase.NextCheck(now, now.AddDate(0, 0, 3))
	if !far.After(medium) || !medium.After(near) {
		t.Fatalf("scheduling does not tighten: far %v medium %v near %v", far, medium, near)
	}
	if !near.After(now) {
		t.Fatal("a check was scheduled in the past")
	}
}
