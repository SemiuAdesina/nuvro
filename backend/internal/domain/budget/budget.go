package budget

import (
	"errors"
	"slices"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/money"
)

var ErrInvalid = errors.New("budget: invalid estimate")
var ErrCurrencyMismatch = errors.New("budget: mixed currencies")

type Category string

const (
	Flight         Category = "flight"
	Stay           Category = "stay"
	Food           Category = "food"
	Activities     Category = "activities"
	LocalTransport Category = "local_transport"
	Entry          Category = "entry"
	Connectivity   Category = "connectivity"
	Other          Category = "other"
	Buffer         Category = "buffer"
)

func Categories() []Category {
	return []Category{Flight, Stay, Food, Activities, LocalTransport, Entry, Connectivity, Other, Buffer}
}

func (c Category) Valid() bool { return slices.Contains(Categories(), c) }

type Confidence string

const (
	High    Confidence = "high"
	Medium  Confidence = "medium"
	Low     Confidence = "low"
	Unknown Confidence = "unknown"
)

func (c Confidence) Valid() bool {
	switch c {
	case High, Medium, Low, Unknown:
		return true
	default:
		return false
	}
}

func (c Confidence) Rank() int {
	switch c {
	case High:
		return 3
	case Medium:
		return 2
	case Low:
		return 1
	default:
		return 0
	}
}

type Range struct {
	Minimum  money.Amount
	Expected money.Amount
	Maximum  money.Amount
}

func NewRange(minimum, expected, maximum money.Amount) (Range, error) {
	if minimum.Currency() != expected.Currency() || expected.Currency() != maximum.Currency() {
		return Range{}, ErrCurrencyMismatch
	}
	if minimum.Minor() > expected.Minor() || expected.Minor() > maximum.Minor() {
		return Range{}, ErrInvalid
	}
	return Range{Minimum: minimum, Expected: expected, Maximum: maximum}, nil
}

func (r Range) Add(other Range) (Range, error) {
	minimum, err := r.Minimum.Add(other.Minimum)
	if err != nil {
		return Range{}, err
	}
	expected, err := r.Expected.Add(other.Expected)
	if err != nil {
		return Range{}, err
	}
	maximum, err := r.Maximum.Add(other.Maximum)
	if err != nil {
		return Range{}, err
	}
	return Range{Minimum: minimum, Expected: expected, Maximum: maximum}, nil
}

type Line struct {
	Category    Category
	Estimated   Range
	HasEstimate bool
	Booked      money.Amount
	HasBooked   bool
	Confidence  Confidence
	Source      string
	ObservedAt  time.Time
}

func (l Line) Validate() error {
	if !l.Category.Valid() || !l.Confidence.Valid() {
		return ErrInvalid
	}
	if l.HasBooked && l.Booked.IsNegative() {
		return ErrInvalid
	}
	return nil
}

func (l Line) Effective() (money.Amount, bool) {
	if l.HasBooked {
		return l.Booked, true
	}
	if l.HasEstimate {
		return l.Estimated.Expected, true
	}
	return money.Amount{}, false
}
