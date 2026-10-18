//go:build integration

package integration_test

import (
	"testing"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/adapter/postgres"
	"github.com/SemiuAdesina/nuvro/backend/internal/apperror"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/commerce"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/money"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/pricewatch"
)

func amountOf(t *testing.T, minor int64) money.Amount {
	t.Helper()
	value, err := money.New(minor, money.Currency("EUR"))
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func handoffFixture(t *testing.T, tripID, userID string) commerce.Handoff {
	t.Helper()
	now := time.Now().UTC().Truncate(time.Second)
	return commerce.Handoff{
		ID: newIdentifier(t), OfferID: "offer-1", TripID: tripID, UserID: userID,
		Partner: "fixture", Product: commerce.Flight,
		DestinationURL: "https://partner.invalid/book/offer-1", Title: "Lisbon to Funchal",
		Price:            amountOf(t, 22_000),
		Disclosure:       commerce.Disclosure{Code: commerce.EligibleBookingDisclosure, Partner: "Fixture Travel"},
		AttributionToken: newIdentifier(t), CreatedAt: now, ExpiresAt: now.Add(30 * time.Minute),
	}
}

func TestHandoffAndClickPersistOnce(t *testing.T) {
	pool, _ := openDatabase(t)
	store := postgres.NewStore(pool.Pool)
	tripID := seedTrip(t, pool, store)
	userID := ownerOf(t, pool, tripID)
	handoff := handoffFixture(t, tripID, userID)

	if err := store.SaveHandoff(t.Context(), handoff); err != nil {
		t.Fatal(err)
	}
	stored, err := store.Handoff(t.Context(), handoff.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.DestinationURL != handoff.DestinationURL || stored.Price.Minor() != 22_000 {
		t.Fatalf("stored = %+v", stored)
	}
	if stored.AttributionToken != handoff.AttributionToken {
		t.Fatal("the attribution token was lost")
	}
	if stored.Disclosure.Code != commerce.EligibleBookingDisclosure {
		t.Fatal("disclosure was lost")
	}

	click := commerce.Click{
		ID: newIdentifier(t), HandoffID: handoff.ID, Partner: handoff.Partner,
		Product: handoff.Product, OccurredAt: time.Now().UTC(),
	}
	if err := store.RecordClick(t.Context(), click); err != nil {
		t.Fatal(err)
	}
	repeat := click
	repeat.ID = newIdentifier(t)
	if err := store.RecordClick(t.Context(), repeat); err != nil {
		t.Fatal(err)
	}
	count, err := store.ClickCount(t.Context(), handoff.ID)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("recorded %d clicks for one handoff", count)
	}
}

func TestDatabaseRejectsUnsafeHandoffDestinations(t *testing.T) {
	pool, _ := openDatabase(t)
	store := postgres.NewStore(pool.Pool)
	tripID := seedTrip(t, pool, store)
	handoff := handoffFixture(t, tripID, ownerOf(t, pool, tripID))
	handoff.DestinationURL = "http://partner.invalid/book"
	if err := store.SaveHandoff(t.Context(), handoff); err == nil {
		t.Fatal("a plain http destination was stored")
	}
}

func TestWebhookEventsAreClaimedOnce(t *testing.T) {
	pool, _ := openDatabase(t)
	store := postgres.NewStore(pool.Pool)
	first, err := store.ClaimWebhookEvent(t.Context(), newIdentifier(t), "fixture", "evt-1", "booking.completed", "booking_completed")
	if err != nil {
		t.Fatal(err)
	}
	if !first {
		t.Fatal("the first delivery was not claimed")
	}
	second, err := store.ClaimWebhookEvent(t.Context(), newIdentifier(t), "fixture", "evt-1", "booking.completed", "booking_completed")
	if err != nil {
		t.Fatal(err)
	}
	if second {
		t.Fatal("a repeated delivery was claimed twice")
	}
	if err := store.MarkWebhookProcessed(t.Context(), "fixture", "evt-1", time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
}

func TestConversionsUpsertOnProviderIdentity(t *testing.T) {
	pool, _ := openDatabase(t)
	store := postgres.NewStore(pool.Pool)
	tripID := seedTrip(t, pool, store)
	handoff := handoffFixture(t, tripID, ownerOf(t, pool, tripID))
	if err := store.SaveHandoff(t.Context(), handoff); err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC().Truncate(time.Second)
	conversion := commerce.Conversion{
		ID: newIdentifier(t), Partner: "fixture", ProviderConversionID: "conv-1",
		HandoffID: handoff.ID, Product: commerce.Flight, Status: commerce.BookingCreated,
		BookingReference: "ABC123XYZ", GrossValue: amountOf(t, 50_000), HasGrossValue: true,
		Commission: amountOf(t, 2_500), HasCommission: true, OccurredAt: now,
	}
	if _, err := store.SaveConversion(t.Context(), conversion); err != nil {
		t.Fatal(err)
	}
	update := conversion
	update.ID = newIdentifier(t)
	update.Status = commerce.BookingCompleted
	if _, err := store.SaveConversion(t.Context(), update); err != nil {
		t.Fatal(err)
	}
	stored, err := store.Conversion(t.Context(), "fixture", "conv-1")
	if err != nil {
		t.Fatal(err)
	}
	if stored.Status != commerce.BookingCompleted {
		t.Fatalf("status = %q", stored.Status)
	}
	if stored.ID != conversion.ID {
		t.Fatal("a second conversion row was created for one provider identity")
	}
	if !stored.HasCommission || stored.Commission.Minor() != 2_500 {
		t.Fatalf("commission lost: %+v", stored)
	}
	if stored.Commission.Currency().String() != "EUR" || stored.GrossValue.Currency().String() != "EUR" {
		t.Fatal("accounting currency was converted")
	}
	if _, err := store.Conversion(t.Context(), "fixture", "never-seen"); apperror.CodeOf(err) != apperror.CodeNotFound {
		t.Fatal("an unknown conversion did not report not found")
	}
}

func TestPriceWatchesPersistAndSchedule(t *testing.T) {
	pool, _ := openDatabase(t)
	store := postgres.NewStore(pool.Pool)
	tripID := seedTrip(t, pool, store)
	userID := ownerOf(t, pool, tripID)
	now := time.Now().UTC().Truncate(time.Second)
	watch := pricewatch.Watch{
		ID: newIdentifier(t), UserID: userID, TripID: tripID, Product: commerce.Flight,
		Target: pricewatch.Target{
			Origin: "LIS", Destination: "FNC", StartDate: now.AddDate(0, 1, 0),
			EndDate: now.AddDate(0, 1, 6), Adults: 2,
		},
		Baseline: amountOf(t, 40_000), Threshold: pricewatch.DefaultThreshold(),
		Status: pricewatch.Active, CreatedAt: now, NextCheckAt: now.Add(-time.Minute),
		ExpiresAt: now.AddDate(0, 1, 0),
	}
	if err := store.SavePriceWatch(t.Context(), watch); err != nil {
		t.Fatal(err)
	}
	stored, err := store.PriceWatch(t.Context(), watch.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.Baseline.Minor() != 40_000 || stored.Target.Origin != "LIS" {
		t.Fatalf("stored = %+v", stored)
	}
	if stored.Threshold.PercentageDrop != watch.Threshold.PercentageDrop {
		t.Fatal("threshold lost")
	}
	count, err := store.ActivePriceWatchCount(t.Context(), userID)
	if err != nil || count != 1 {
		t.Fatalf("%d active watches %v", count, err)
	}
	due, err := store.DuePriceWatches(t.Context(), now, 50)
	if err != nil {
		t.Fatal(err)
	}
	if !containsWatch(due, watch.ID) {
		t.Fatal("a due watch was not returned")
	}
	if err := store.RecordObservation(t.Context(), newIdentifier(t), pricewatch.Observation{
		WatchID: watch.ID, Price: amountOf(t, 32_000), Available: true, ObservedAt: now,
	}); err != nil {
		t.Fatal(err)
	}
	if err := store.ReschedulePriceWatch(t.Context(), watch.ID, now, now.Add(6*time.Hour), 32_000); err != nil {
		t.Fatal(err)
	}
	rescheduled, err := store.PriceWatch(t.Context(), watch.ID)
	if err != nil {
		t.Fatal(err)
	}
	if rescheduled.Baseline.Minor() != 32_000 || !rescheduled.NextCheckAt.After(now) {
		t.Fatalf("rescheduled = %+v", rescheduled)
	}
	stillDue, err := store.DuePriceWatches(t.Context(), now, 50)
	if err != nil {
		t.Fatal(err)
	}
	if containsWatch(stillDue, watch.ID) {
		t.Fatal("a rescheduled watch is still due")
	}
}

func TestExpiredWatchesAreEnded(t *testing.T) {
	pool, _ := openDatabase(t)
	store := postgres.NewStore(pool.Pool)
	tripID := seedTrip(t, pool, store)
	userID := ownerOf(t, pool, tripID)
	now := time.Now().UTC().Truncate(time.Second)
	watch := pricewatch.Watch{
		ID: newIdentifier(t), UserID: userID, TripID: tripID, Product: commerce.Stay,
		Target: pricewatch.Target{
			PropertyID: "hotel-1", StartDate: now.AddDate(0, 0, 1), EndDate: now.AddDate(0, 0, 4), Adults: 2,
		},
		Baseline: amountOf(t, 70_000), Threshold: pricewatch.DefaultThreshold(),
		Status: pricewatch.Active, CreatedAt: now, ExpiresAt: now.Add(time.Minute),
	}
	if err := store.SavePriceWatch(t.Context(), watch); err != nil {
		t.Fatal(err)
	}
	afterExpiry := now.Add(2 * time.Minute)
	if err := store.ExpirePriceWatches(t.Context(), afterExpiry); err != nil {
		t.Fatal(err)
	}
	stored, err := store.PriceWatch(t.Context(), watch.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.Status != pricewatch.Ended {
		t.Fatalf("status = %q", stored.Status)
	}
	due, err := store.DuePriceWatches(t.Context(), afterExpiry, 50)
	if err != nil {
		t.Fatal(err)
	}
	if containsWatch(due, watch.ID) {
		t.Fatal("an ended watch is still due")
	}
}

func TestDatabaseRejectsInventedCommerceVocabulary(t *testing.T) {
	pool, _ := openDatabase(t)
	store := postgres.NewStore(pool.Pool)
	tripID := seedTrip(t, pool, store)
	handoff := handoffFixture(t, tripID, ownerOf(t, pool, tripID))
	handoff.Product = "yacht"
	if err := store.SaveHandoff(t.Context(), handoff); err == nil {
		t.Fatal("an invented product type was stored")
	}
}

func containsWatch(watches []pricewatch.Watch, watchID string) bool {
	for _, watch := range watches {
		if watch.ID == watchID {
			return true
		}
	}
	return false
}
