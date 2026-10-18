package discover

import (
	"context"

	"golang.org/x/sync/errgroup"

	"github.com/SemiuAdesina/nuvro/backend/internal/apperror"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/destination"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/discovery"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
	"github.com/SemiuAdesina/nuvro/backend/internal/engine/recommendation"
	"github.com/SemiuAdesina/nuvro/backend/internal/engine/understanding"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
)

const catalogLimit = 200
const evaluationConcurrency = 8

type Ranked = recommendation.Result

type Engines struct {
	Understanding Understanding
	Feasibility   Feasibility
	Matching      Matching
	Ranking       Ranking
}

type Service struct {
	profiles        ProfileReader
	catalog         DestinationCatalog
	engines         Engines
	personalization Personalization
	timer           clock.Clock
}

func NewService(profiles ProfileReader, catalog DestinationCatalog, engines Engines, timer clock.Clock) *Service {
	return &Service{profiles: profiles, catalog: catalog, engines: engines, timer: timer}
}

func (s *Service) WithPersonalization(context Personalization) *Service {
	s.personalization = context
	return s
}

func (s *Service) Matches(ctx context.Context, principal user.Principal, request Request) (Response, error) {
	if principal.UserID == "" {
		return Response{}, apperror.New(apperror.CodeUnauthenticated, "Authentication is required.")
	}
	intent, err := s.intent(ctx, principal, request)
	if err != nil {
		return Response{}, err
	}
	catalog, err := s.catalog.Destinations(ctx, catalogLimit)
	if err != nil {
		return Response{}, err
	}
	candidates, err := s.evaluate(ctx, intent, catalog)
	if err != nil {
		return Response{}, err
	}
	return Response{
		Intent:          intent,
		Result:          s.engines.Ranking.Rank(ctx, candidates),
		MissingContext:  intent.MissingContext(),
		GeneratedAt:     s.timer.Now().UTC(),
		EvidenceVersion: s.engines.Matching.Version(),
	}, nil
}

func (s *Service) intent(ctx context.Context, principal user.Principal, request Request) (discovery.Intent, error) {
	profile, err := s.profiles.Profile(ctx, principal.UserID)
	if err != nil {
		return discovery.Intent{}, err
	}
	preferences, err := s.profiles.Preferences(ctx, principal.UserID)
	if err != nil {
		return discovery.Intent{}, err
	}
	base := profileIntent(profile, preferences)
	base = s.personalize(ctx, principal, base)
	if request.Overrides.Budget == nil {
		return discovery.Intent{}, apperror.New(apperror.CodeInvalidInput, "A trip budget is required.")
	}
	intent, err := s.engines.Understanding.Understand(ctx, understanding.Input{
		Profile: base, Supplied: request.Overrides, Party: request.Party,
		Window: request.Window, Query: request.Query, Now: s.timer.Now(),
	})
	if err != nil {
		return discovery.Intent{}, apperror.New(apperror.CodeInvalidInput, "The discovery request is not valid.")
	}
	return intent, nil
}

func (s *Service) evaluate(ctx context.Context, intent discovery.Intent, catalog []destination.Destination) ([]recommendation.Candidate, error) {
	candidates := make([]recommendation.Candidate, len(catalog))
	group, groupCtx := errgroup.WithContext(ctx)
	group.SetLimit(evaluationConcurrency)
	for index, target := range catalog {
		group.Go(func() error {
			result := s.engines.Feasibility.Evaluate(groupCtx, intent, target)
			score, err := s.engines.Matching.Score(intent, target, result, s.engines.Matching.Version())
			if err != nil {
				return apperror.Wrap(apperror.CodeInternal, "", err)
			}
			candidates[index] = recommendation.Candidate{Destination: target, Match: score, Feasibility: result}
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		return nil, err
	}
	return candidates, nil
}
