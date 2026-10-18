package language

import (
	"context"
	"errors"
	"strings"

	domain "github.com/SemiuAdesina/nuvro/backend/internal/domain/language"
)

const MaximumTextBytes = 5000
const maximumContextBytes = 200

var ErrUnavailable = errors.New("language: translation provider unavailable")
var ErrUnsupported = errors.New("language: unsupported language")
var ErrInvalidText = errors.New("language: text is empty or too long")

type Engine struct{ provider Provider }

func New(provider Provider) *Engine { return &Engine{provider: provider} }

func (e *Engine) Configured() bool { return e.provider != nil }

type Request struct {
	Source  string
	Target  string
	Text    string
	Context string
}

func (e *Engine) Translate(ctx context.Context, request Request) (domain.Translation, error) {
	target, err := domain.Parse(request.Target)
	if err != nil {
		return domain.Translation{}, ErrUnsupported
	}
	source := domain.Undetermined
	detected := true
	if strings.TrimSpace(request.Source) != "" {
		source, err = domain.Parse(request.Source)
		if err != nil {
			return domain.Translation{}, ErrUnsupported
		}
		detected = false
	}
	text := strings.TrimSpace(request.Text)
	if text == "" || len(text) > MaximumTextBytes {
		return domain.Translation{}, ErrInvalidText
	}
	if e.provider == nil {
		return domain.Translation{}, ErrUnavailable
	}
	if !e.provider.Supports(target) || (!detected && !e.provider.Supports(source)) {
		return domain.Translation{}, ErrUnsupported
	}
	result, err := e.provider.Translate(ctx, Query{
		Source: source, Target: target, Text: text, Context: trimContext(request.Context),
	})
	if err != nil {
		return domain.Translation{}, ErrUnavailable
	}
	return e.translation(source, target, text, detected, result)
}

func (e *Engine) translation(source, target domain.Code, text string, detected bool, result Result) (domain.Translation, error) {
	translated := strings.TrimSpace(result.Translated)
	if translated == "" || len(translated) > MaximumTextBytes*2 {
		return domain.Translation{}, ErrUnavailable
	}
	resolved := source
	if detected {
		if result.Source.Valid() {
			resolved = result.Source
		} else {
			return domain.Translation{}, ErrUnavailable
		}
	}
	return domain.Translation{
		Source: resolved, Target: target, SourceText: text, Translated: translated,
		Detected: detected, Confidence: confidenceOrUnknown(result.Confidence), Provider: result.Provider,
	}, nil
}

func trimContext(value string) string {
	trimmed := strings.TrimSpace(value)
	if len(trimmed) > maximumContextBytes {
		return trimmed[:maximumContextBytes]
	}
	return trimmed
}

func confidenceOrUnknown(value string) string {
	switch value {
	case "high", "medium", "low":
		return value
	default:
		return "unknown"
	}
}
