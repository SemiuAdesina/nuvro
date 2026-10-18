package money_test

import (
	"errors"
	"math"
	"testing"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/money"
)

func TestParseCurrencyAcceptsIsoAlphabeticCode(t *testing.T) {
	parsed, err := money.ParseCurrency("USD")
	if err != nil {
		t.Fatalf("ParseCurrency returned error: %v", err)
	}
	if parsed.String() != "USD" {
		t.Fatalf("got %q, want %q", parsed, "USD")
	}
}

func TestParseCurrencyRejectsMalformedCode(t *testing.T) {
	for _, raw := range []string{"", "us", "usd", "USDD", "US1", "US "} {
		if _, err := money.ParseCurrency(raw); !errors.Is(err, money.ErrInvalidCurrency) {
			t.Errorf("ParseCurrency(%q) error = %v, want ErrInvalidCurrency", raw, err)
		}
	}
}

func TestNewStoresMinorUnitsExactly(t *testing.T) {
	amount, err := money.New(12550, "USD")
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	if amount.Minor() != 12550 {
		t.Fatalf("Minor() = %d, want 12550", amount.Minor())
	}
	if amount.Currency() != "USD" {
		t.Fatalf("Currency() = %q, want USD", amount.Currency())
	}
	if amount.String() != "12550 USD" {
		t.Fatalf("String() = %q, want %q", amount.String(), "12550 USD")
	}
}

func TestNewRejectsMalformedCurrency(t *testing.T) {
	if _, err := money.New(1, "dollars"); !errors.Is(err, money.ErrInvalidCurrency) {
		t.Fatalf("error = %v, want ErrInvalidCurrency", err)
	}
}

func TestAddSumsMatchingCurrencies(t *testing.T) {
	left := mustAmount(t, 2500, "EUR")
	right := mustAmount(t, 799, "EUR")

	sum, err := left.Add(right)
	if err != nil {
		t.Fatalf("Add returned error: %v", err)
	}
	if sum.Minor() != 3299 {
		t.Fatalf("Minor() = %d, want 3299", sum.Minor())
	}
}

func TestAddRejectsCurrencyMismatch(t *testing.T) {
	left := mustAmount(t, 100, "EUR")
	right := mustAmount(t, 100, "USD")

	if _, err := left.Add(right); !errors.Is(err, money.ErrCurrencyMismatch) {
		t.Fatalf("error = %v, want ErrCurrencyMismatch", err)
	}
}

func TestAddDetectsOverflow(t *testing.T) {
	left := mustAmount(t, math.MaxInt64, "JPY")
	right := mustAmount(t, 1, "JPY")

	if _, err := left.Add(right); !errors.Is(err, money.ErrOverflow) {
		t.Fatalf("error = %v, want ErrOverflow", err)
	}
}

func TestSubtractProducesNegativeBalance(t *testing.T) {
	left := mustAmount(t, 1000, "GBP")
	right := mustAmount(t, 2500, "GBP")

	difference, err := left.Subtract(right)
	if err != nil {
		t.Fatalf("Subtract returned error: %v", err)
	}
	if difference.Minor() != -1500 {
		t.Fatalf("Minor() = %d, want -1500", difference.Minor())
	}
	if !difference.IsNegative() {
		t.Fatal("IsNegative() = false, want true")
	}
}

func TestSubtractDetectsOverflowOnMinInt64Operand(t *testing.T) {
	left := mustAmount(t, 0, "JPY")
	right := mustAmount(t, math.MinInt64, "JPY")

	if _, err := left.Subtract(right); !errors.Is(err, money.ErrOverflow) {
		t.Fatalf("error = %v, want ErrOverflow", err)
	}
}

func TestCompareOrdersAmountsOfOneCurrency(t *testing.T) {
	smaller := mustAmount(t, 100, "NGN")
	larger := mustAmount(t, 900, "NGN")

	assertCompare(t, smaller, larger, -1)
	assertCompare(t, larger, smaller, 1)
	assertCompare(t, smaller, smaller, 0)
}

func TestCompareRejectsCurrencyMismatch(t *testing.T) {
	if _, err := mustAmount(t, 1, "NGN").Compare(mustAmount(t, 1, "USD")); !errors.Is(err, money.ErrCurrencyMismatch) {
		t.Fatalf("error = %v, want ErrCurrencyMismatch", err)
	}
}

func TestZeroAmountIsZero(t *testing.T) {
	if !mustAmount(t, 0, "USD").IsZero() {
		t.Fatal("IsZero() = false, want true")
	}
	if mustAmount(t, 1, "USD").IsZero() {
		t.Fatal("IsZero() = true, want false")
	}
}

func assertCompare(t *testing.T, left money.Amount, right money.Amount, want int) {
	t.Helper()
	got, err := left.Compare(right)
	if err != nil {
		t.Fatalf("Compare returned error: %v", err)
	}
	if got != want {
		t.Fatalf("Compare(%s, %s) = %d, want %d", left, right, got, want)
	}
}

func mustAmount(t *testing.T, minor int64, currency money.Currency) money.Amount {
	t.Helper()
	amount, err := money.New(minor, currency)
	if err != nil {
		t.Fatalf("New(%d, %q) returned error: %v", minor, currency, err)
	}
	return amount
}

func FuzzParseCurrency(f *testing.F) {
	f.Add("USD")
	f.Add("usd")
	f.Add("")
	f.Add("EURO")
	f.Add("€€€")
	f.Fuzz(func(t *testing.T, raw string) {
		currency, err := money.ParseCurrency(raw)
		if err != nil {
			if currency != "" {
				t.Fatal("rejected currency leaked a value")
			}
			return
		}
		if len(currency) != 3 || currency.String() != raw {
			t.Fatalf("accepted %q as %q", raw, currency)
		}
		for _, r := range currency {
			if r < 'A' || r > 'Z' {
				t.Fatalf("accepted non-uppercase %q", raw)
			}
		}
	})
}
