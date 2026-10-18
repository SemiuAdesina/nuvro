package money

import (
	"errors"
	"fmt"
	"time"
)

const rateScale = 1_000_000

var (
	ErrNonPositiveRate  = errors.New("money: conversion rate must be positive")
	ErrMissingRateTime  = errors.New("money: conversion requires a rate timestamp")
	ErrMissingRateOwner = errors.New("money: conversion requires a rate source")
	ErrSameCurrency     = errors.New("money: conversion requires distinct currencies")
)

type Rate struct {
	micros int64
}

func NewRate(micros int64) (Rate, error) {
	if micros <= 0 {
		return Rate{}, ErrNonPositiveRate
	}
	return Rate{micros: micros}, nil
}

func (r Rate) Micros() int64 {
	return r.micros
}

func (r Rate) Apply(source Amount) (int64, error) {
	product := source.Minor() * r.micros
	if source.Minor() != 0 && product/source.Minor() != r.micros {
		return 0, ErrOverflow
	}
	half := int64(rateScale / 2)
	if product < 0 {
		return (product - half) / rateScale, nil
	}
	return (product + half) / rateScale, nil
}

type Conversion struct {
	Source     Amount
	Target     Amount
	Rate       Rate
	ObservedAt time.Time
	Provider   string
}

func NewConversion(source Amount, target Amount, rate Rate, observedAt time.Time, provider string) (Conversion, error) {
	if source.Currency() == target.Currency() {
		return Conversion{}, fmt.Errorf("%w: %s", ErrSameCurrency, source.Currency())
	}
	if observedAt.IsZero() {
		return Conversion{}, ErrMissingRateTime
	}
	if provider == "" {
		return Conversion{}, ErrMissingRateOwner
	}
	return Conversion{
		Source:     source,
		Target:     target,
		Rate:       rate,
		ObservedAt: observedAt.UTC(),
		Provider:   provider,
	}, nil
}

func (c Conversion) Age(now time.Time) time.Duration {
	return now.UTC().Sub(c.ObservedAt)
}
