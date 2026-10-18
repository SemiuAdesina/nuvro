package language

import (
	"context"

	domain "github.com/SemiuAdesina/nuvro/backend/internal/domain/language"
)

type Query struct {
	Source  domain.Code
	Target  domain.Code
	Text    string
	Context string
}

type Result struct {
	Source     domain.Code
	Translated string
	Confidence string
	Provider   string
}

type Provider interface {
	Translate(context.Context, Query) (Result, error)
	Supports(domain.Code) bool
}
