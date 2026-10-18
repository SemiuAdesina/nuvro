package money

import (
	"errors"
	"fmt"
	"math"
)

var (
	ErrInvalidCurrency  = errors.New("money: currency must be three uppercase letters")
	ErrCurrencyMismatch = errors.New("money: currency mismatch")
	ErrOverflow         = errors.New("money: amount overflows int64 minor units")
)

type Currency string

func ParseCurrency(raw string) (Currency, error) {
	if len(raw) != 3 {
		return "", fmt.Errorf("%w: %q", ErrInvalidCurrency, raw)
	}
	for _, r := range raw {
		if r < 'A' || r > 'Z' {
			return "", fmt.Errorf("%w: %q", ErrInvalidCurrency, raw)
		}
	}
	return Currency(raw), nil
}

func (c Currency) String() string {
	return string(c)
}

type Amount struct {
	minor    int64
	currency Currency
}

func New(minor int64, currency Currency) (Amount, error) {
	parsed, err := ParseCurrency(string(currency))
	if err != nil {
		return Amount{}, err
	}
	return Amount{minor: minor, currency: parsed}, nil
}

func (a Amount) Minor() int64 {
	return a.minor
}

func (a Amount) Currency() Currency {
	return a.currency
}

func (a Amount) IsZero() bool {
	return a.minor == 0
}

func (a Amount) IsNegative() bool {
	return a.minor < 0
}

func (a Amount) Add(other Amount) (Amount, error) {
	if err := a.assertSameCurrency(other); err != nil {
		return Amount{}, err
	}
	sum := a.minor + other.minor
	if (a.minor > 0 && other.minor > 0 && sum < 0) || (a.minor < 0 && other.minor < 0 && sum >= 0) {
		return Amount{}, ErrOverflow
	}
	return Amount{minor: sum, currency: a.currency}, nil
}

func (a Amount) Subtract(other Amount) (Amount, error) {
	if err := a.assertSameCurrency(other); err != nil {
		return Amount{}, err
	}
	negated, err := other.negate()
	if err != nil {
		return Amount{}, err
	}
	return a.Add(negated)
}

func (a Amount) Compare(other Amount) (int, error) {
	if err := a.assertSameCurrency(other); err != nil {
		return 0, err
	}
	switch {
	case a.minor < other.minor:
		return -1, nil
	case a.minor > other.minor:
		return 1, nil
	default:
		return 0, nil
	}
}

func (a Amount) String() string {
	return fmt.Sprintf("%d %s", a.minor, a.currency)
}

func (a Amount) negate() (Amount, error) {
	if a.minor == math.MinInt64 {
		return Amount{}, ErrOverflow
	}
	return Amount{minor: -a.minor, currency: a.currency}, nil
}

func (a Amount) assertSameCurrency(other Amount) error {
	if a.currency != other.currency {
		return fmt.Errorf("%w: %s and %s", ErrCurrencyMismatch, a.currency, other.currency)
	}
	return nil
}
