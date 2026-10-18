package alert_test

import (
	"testing"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/alert"
)

var now = time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC)

func valid() alert.Alert {
	return alert.Alert{
		ID: "alert", TripID: "trip", Type: alert.WeatherImpact, Severity: alert.Attention,
		TitleCode: "alert.weather_impact", MessageData: map[string]string{"activity": "Garden"},
		Action: alert.AdjustPlan, Status: alert.Unread, Fingerprint: "abcd1234abcd1234",
		CreatedAt: now, EffectiveAt: now, ExpiresAt: now.Add(3 * time.Hour),
	}
}

func TestAlertValidation(t *testing.T) {
	if err := valid().Validate(); err != nil {
		t.Fatal(err)
	}
	for name, mutate := range map[string]func(*alert.Alert){
		"no trip":             func(a *alert.Alert) { a.TripID = "" },
		"no title code":       func(a *alert.Alert) { a.TitleCode = "" },
		"invented type":       func(a *alert.Alert) { a.Type = "marketing" },
		"invented severity":   func(a *alert.Alert) { a.Severity = "critical" },
		"invented status":     func(a *alert.Alert) { a.Status = "snoozed" },
		"invented action":     func(a *alert.Alert) { a.Action = "buy_now" },
		"no fingerprint":      func(a *alert.Alert) { a.Fingerprint = "" },
		"no effective time":   func(a *alert.Alert) { a.EffectiveAt = time.Time{} },
		"expiry before start": func(a *alert.Alert) { a.ExpiresAt = a.EffectiveAt.Add(-time.Hour) },
		"oversized data":      func(a *alert.Alert) { a.MessageData = map[string]string{"k": string(make([]byte, 300))} },
	} {
		value := valid()
		mutate(&value)
		if value.Validate() == nil {
			t.Fatalf("%s was accepted", name)
		}
	}
}

func TestSeverityVocabularyExcludesCritical(t *testing.T) {
	if alert.Severity("critical").Valid() {
		t.Fatal("critical is not a NUVRO severity")
	}
	for _, severity := range []alert.Severity{alert.Info, alert.Attention, alert.Important} {
		if !severity.Valid() {
			t.Fatalf("%q was rejected", severity)
		}
	}
}

func TestStatusLifecycle(t *testing.T) {
	if !alert.Unread.CanTransitionTo(alert.Read) || !alert.Read.CanTransitionTo(alert.Resolved) {
		t.Fatal("a legitimate transition was refused")
	}
	if alert.Read.CanTransitionTo(alert.Unread) || alert.Resolved.CanTransitionTo(alert.Read) {
		t.Fatal("an alert was reopened")
	}
	if alert.Unread.CanTransitionTo(alert.Expired) {
		t.Fatal("a client can mark an alert expired")
	}
	if alert.Unread.CanTransitionTo("snoozed") {
		t.Fatal("an invented status was accepted")
	}
}

func TestVisibilityFollowsStatusAndTime(t *testing.T) {
	value := valid()
	if !value.Visible(now) {
		t.Fatal("a fresh unread alert is not visible")
	}
	if value.Visible(now.Add(4 * time.Hour)) {
		t.Fatal("an expired alert is still visible")
	}
	future := valid()
	future.EffectiveAt = now.Add(time.Hour)
	if future.Visible(now) {
		t.Fatal("an alert became visible before it takes effect")
	}
	resolved := valid()
	resolved.Status = alert.Resolved
	if resolved.Visible(now) {
		t.Fatal("a resolved alert is still visible")
	}
	permanent := valid()
	permanent.ExpiresAt = time.Time{}
	if !permanent.Visible(now.Add(100*time.Hour)) || permanent.ExpiredAt(now) {
		t.Fatal("an alert without an expiry was treated as expired")
	}
}

func TestFingerprintIsStableAndDiscriminating(t *testing.T) {
	first := alert.Fingerprint("trip", alert.WeatherImpact, "item", "2026-10-16T15")
	if first != alert.Fingerprint("trip", alert.WeatherImpact, "item", "2026-10-16T15") {
		t.Fatal("the same inputs produced different fingerprints")
	}
	for name, other := range map[string]string{
		"other trip":     alert.Fingerprint("other", alert.WeatherImpact, "item", "2026-10-16T15"),
		"other type":     alert.Fingerprint("trip", alert.LeaveReminder, "item", "2026-10-16T15"),
		"other resource": alert.Fingerprint("trip", alert.WeatherImpact, "other", "2026-10-16T15"),
		"other window":   alert.Fingerprint("trip", alert.WeatherImpact, "item", "2026-10-16T16"),
	} {
		if other == first {
			t.Fatalf("%s collided with the original fingerprint", name)
		}
	}
}
