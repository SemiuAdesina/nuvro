package language

import (
	"context"
	"errors"

	"github.com/SemiuAdesina/nuvro/backend/internal/apperror"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/destination"
	domain "github.com/SemiuAdesina/nuvro/backend/internal/domain/language"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/trip"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
	engine "github.com/SemiuAdesina/nuvro/backend/internal/engine/language"
)

type Trips interface {
	Read(context.Context, user.Principal, string) (trip.Trip, error)
}

type Catalog interface {
	Destination(context.Context, string) (destination.Destination, error)
}

type Profiles interface {
	Profile(context.Context, string) (user.Profile, error)
}

type Translator interface {
	Translate(context.Context, engine.Request) (domain.Translation, error)
	Phrases(ctx context.Context, source, target domain.Code, categories []domain.Category) domain.Pack
	Configured() bool
}

type Service struct {
	trips      Trips
	catalog    Catalog
	profiles   Profiles
	translator Translator
}

func NewService(trips Trips, catalog Catalog, profiles Profiles, translator Translator) *Service {
	return &Service{trips: trips, catalog: catalog, profiles: profiles, translator: translator}
}

type TranslateRequest struct {
	TripID  string
	Source  string
	Target  string
	Text    string
	Context string
}

func (s *Service) Translate(ctx context.Context, principal user.Principal, request TranslateRequest) (domain.Translation, error) {
	source, target, err := s.resolve(ctx, principal, request.TripID, request.Source, request.Target)
	if err != nil {
		return domain.Translation{}, err
	}
	result, err := s.translator.Translate(ctx, engine.Request{
		Source: source.String(), Target: target.String(), Text: request.Text, Context: request.Context,
	})
	if err != nil {
		return domain.Translation{}, translationError(err)
	}
	return result, nil
}

type PhraseRequest struct {
	TripID     string
	Target     string
	Categories []string
}

func (s *Service) Phrases(ctx context.Context, principal user.Principal, request PhraseRequest) (domain.Pack, error) {
	source, target, err := s.resolve(ctx, principal, request.TripID, "", request.Target)
	if err != nil {
		return domain.Pack{}, err
	}
	categories := make([]domain.Category, 0, len(request.Categories))
	for _, raw := range request.Categories {
		category, valid := domain.ParseCategory(raw)
		if !valid {
			return domain.Pack{}, apperror.New(apperror.CodeInvalidInput, "That phrase category is not recognised.")
		}
		categories = append(categories, category)
	}
	return s.translator.Phrases(ctx, source, target, categories), nil
}

func (s *Service) resolve(ctx context.Context, principal user.Principal, tripID, requestedSource, requestedTarget string) (domain.Code, domain.Code, error) {
	record, err := s.trips.Read(ctx, principal, tripID)
	if err != nil {
		return "", "", err
	}
	context := engine.Context{AppLanguage: s.appLanguage(ctx, principal)}
	if record.DestinationID != nil {
		if place, lookupErr := s.catalog.Destination(ctx, *record.DestinationID); lookupErr == nil {
			context.DestinationRegion = place.CountryCode
		}
	}
	if requestedTarget != "" {
		target, parseErr := domain.Parse(requestedTarget)
		if parseErr != nil {
			return "", "", apperror.New(apperror.CodeInvalidInput, "That language is not recognised.")
		}
		context.Target = target
		context.HasTarget = true
	}
	source, target := engine.Resolve(context)
	if requestedSource != "" {
		parsed, parseErr := domain.Parse(requestedSource)
		if parseErr != nil {
			return "", "", apperror.New(apperror.CodeInvalidInput, "That language is not recognised.")
		}
		source = parsed
	}
	return source, target, nil
}

func (s *Service) appLanguage(ctx context.Context, principal user.Principal) domain.Code {
	if s.profiles == nil {
		return domain.Code("en")
	}
	profile, err := s.profiles.Profile(ctx, principal.UserID)
	if err != nil {
		return domain.Code("en")
	}
	code, parseErr := domain.Parse(profile.PreferredLanguage)
	if parseErr != nil {
		return domain.Code("en")
	}
	return code
}

func translationError(err error) error {
	switch {
	case errors.Is(err, engine.ErrUnsupported):
		return apperror.New(apperror.CodeInvalidInput, "That language pair is not supported.")
	case errors.Is(err, engine.ErrInvalidText):
		return apperror.New(apperror.CodeInvalidInput, "Provide text to translate, under five thousand characters.")
	default:
		return apperror.New(apperror.CodeUnavailable, "Translation is unavailable.")
	}
}
