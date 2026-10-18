package commerce_test

import (
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/commerce"
)

func partner() commerce.Partner {
	return commerce.Partner{
		Code: "fixture", DisplayName: "Fixture Travel", Status: commerce.Enabled,
		Products: []commerce.ProductType{commerce.Flight}, AllowedDomains: []string{"partner.invalid"},
	}
}

func TestPartnerValidation(t *testing.T) {
	if err := partner().Validate(); err != nil {
		t.Fatal(err)
	}
	for name, mutate := range map[string]func(*commerce.Partner){
		"no code":          func(p *commerce.Partner) { p.Code = "" },
		"no name":          func(p *commerce.Partner) { p.DisplayName = "" },
		"invented status":  func(p *commerce.Partner) { p.Status = "maybe" },
		"no products":      func(p *commerce.Partner) { p.Products = nil },
		"invented product": func(p *commerce.Partner) { p.Products = []commerce.ProductType{"yacht"} },
		"no domains":       func(p *commerce.Partner) { p.AllowedDomains = nil },
		"url as domain":    func(p *commerce.Partner) { p.AllowedDomains = []string{"https://partner.invalid/"} },
	} {
		value := partner()
		mutate(&value)
		if value.Validate() == nil {
			t.Fatalf("%s was accepted", name)
		}
	}
}

func TestPartnerStatusControlsTraffic(t *testing.T) {
	if !commerce.Enabled.Usable() || !commerce.Degraded.Usable() {
		t.Fatal("a usable partner was refused")
	}
	if commerce.Disabled.Usable() {
		t.Fatal("a disabled partner was treated as usable")
	}
}

func TestDestinationAllowlist(t *testing.T) {
	subject := partner()
	for _, allowed := range []string{
		"https://partner.invalid/book",
		"https://partner.invalid/book?ref=abc",
		"https://book.partner.invalid/offer",
		"https://PARTNER.invalid/book",
	} {
		if err := subject.AllowsDestination(allowed); err != nil {
			t.Fatalf("%q was refused: %v", allowed, err)
		}
	}
	for _, blocked := range []string{
		"https://evil.invalid/book",
		"https://partner.invalid.evil.invalid/book",
		"https://notpartner.invalid/book",
		"http://partner.invalid/book",
		"ftp://partner.invalid/book",
		"javascript:alert(1)",
		"//partner.invalid/book",
		"https://user:pass@partner.invalid/book",
		"",
	} {
		if err := subject.AllowsDestination(blocked); err == nil {
			t.Fatalf("%q was allowed", blocked)
		}
	}
}

func FuzzDestinationAllowlist(f *testing.F) {
	f.Add("https://partner.invalid/book")
	f.Add("https://evil.invalid")
	f.Add("")
	f.Add(strings.Repeat("x", 200))
	f.Fuzz(func(t *testing.T, raw string) {
		if partner().AllowsDestination(raw) != nil {
			return
		}
		parsed, err := url.Parse(raw)
		if err != nil {
			t.Fatalf("accepted an unparseable destination: %q", raw)
		}
		if parsed.Scheme != "https" {
			t.Fatalf("accepted a non-https destination: %q", raw)
		}
		if parsed.User != nil {
			t.Fatalf("accepted a destination carrying credentials: %q", raw)
		}
		host := strings.ToLower(parsed.Hostname())
		if host != "partner.invalid" && !strings.HasSuffix(host, ".partner.invalid") {
			t.Fatalf("accepted a destination outside the allowlist: %q", raw)
		}
	})
}

func TestOfferFreshnessDegradesWithAge(t *testing.T) {
	now := time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC)
	future := now.Add(time.Hour)
	for name, test := range map[string]struct {
		observedAt time.Time
		expiresAt  time.Time
		want       commerce.Freshness
	}{
		"just observed": {now, future, commerce.Live},
		"a few minutes": {now.Add(-3 * time.Minute), future, commerce.Live},
		"quarter hour":  {now.Add(-15 * time.Minute), future, commerce.Recent},
		"an hour":       {now.Add(-time.Hour), future, commerce.Stale},
		"past expiry":   {now, now.Add(-time.Minute), commerce.Expired},
	} {
		got := commerce.FreshnessAt(test.observedAt, test.expiresAt, now)
		if got != test.want {
			t.Fatalf("%s = %q, want %q", name, got, test.want)
		}
	}
	if !commerce.Live.Bookable() || !commerce.Recent.Bookable() {
		t.Fatal("a fresh offer was treated as unbookable")
	}
	if commerce.Stale.Bookable() || commerce.Expired.Bookable() {
		t.Fatal("a stale offer was treated as bookable")
	}
}

func TestBookingReferencesAreMasked(t *testing.T) {
	if commerce.MaskReference("ABC123XYZ") != "****3XYZ" {
		t.Fatalf("masked = %q", commerce.MaskReference("ABC123XYZ"))
	}
	if commerce.MaskReference("AB") != "****" {
		t.Fatal("a short reference leaked")
	}
}

func TestConversionValidationAndStatuses(t *testing.T) {
	valid := commerce.Conversion{
		ID: "conversion", Partner: "fixture", ProviderConversionID: "provider-1",
		Product: commerce.Flight, Status: commerce.BookingCompleted,
		OccurredAt: time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC),
	}
	if err := valid.Validate(); err != nil {
		t.Fatal(err)
	}
	for name, mutate := range map[string]func(*commerce.Conversion){
		"no partner":       func(c *commerce.Conversion) { c.Partner = "" },
		"no provider id":   func(c *commerce.Conversion) { c.ProviderConversionID = "" },
		"invented status":  func(c *commerce.Conversion) { c.Status = "maybe_paid" },
		"invented product": func(c *commerce.Conversion) { c.Product = "yacht" },
		"no occurrence":    func(c *commerce.Conversion) { c.OccurredAt = time.Time{} },
	} {
		broken := valid
		mutate(&broken)
		if broken.Validate() == nil {
			t.Fatalf("%s was accepted", name)
		}
	}
	if !commerce.BookingCompleted.ConfirmsBooking() || commerce.CommissionPending.ConfirmsBooking() {
		t.Fatal("booking confirmation is wrong")
	}
	if !commerce.Refunded.EndsBooking() || commerce.BookingCreated.EndsBooking() {
		t.Fatal("booking termination is wrong")
	}
}
