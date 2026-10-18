package app

import (
	"context"

	"github.com/SemiuAdesina/nuvro/backend/internal/adapter/postgres"
	domainlanguage "github.com/SemiuAdesina/nuvro/backend/internal/domain/language"
	"github.com/SemiuAdesina/nuvro/backend/internal/engine/alerting"
	languageengine "github.com/SemiuAdesina/nuvro/backend/internal/engine/language"
	"github.com/SemiuAdesina/nuvro/backend/internal/engine/replanning"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/id"
	"github.com/SemiuAdesina/nuvro/backend/internal/transport/http/authentication"
	alertsusecase "github.com/SemiuAdesina/nuvro/backend/internal/usecase/alerts"
	languageusecase "github.com/SemiuAdesina/nuvro/backend/internal/usecase/language"
	travelusecase "github.com/SemiuAdesina/nuvro/backend/internal/usecase/travel"
)

type runtimeServices struct {
	travel   *travelusecase.Service
	alerts   *alertsusecase.Service
	language *languageusecase.Service
	phrases  travelusecase.PhraseSource
	alerting *alerting.Engine
}

func newRuntimeServices(store *postgres.Store, brains intelligence, access tripAccess, timer clock.Clock) runtimeServices {
	planner := replanning.New(placeReader{engine: brains.places}, routeComputer{engine: brains.routes},
		brains.weather, nil, timer)
	translator := languageengine.New(nil)
	languageService := languageusecase.NewService(access, store, store, translator)
	return runtimeServices{
		travel: travelusecase.NewService(travelusecase.Dependencies{
			Trips: access, Catalog: store, Itineraries: store, Alerts: store, Proposals: store,
			Forecaster: brains.weather, Routes: routeComputer{engine: brains.routes},
			Planner: planner, Preferences: store, Identifiers: id.NewUUID(),
		}, timer),
		alerts:   alertsusecase.NewService(access, store, id.NewUUID(), timer),
		language: languageService,
		phrases:  phraseSource{service: languageService},
		alerting: alerting.New(),
	}
}

type phraseSource struct{ service *languageusecase.Service }

func (p phraseSource) Phrases(ctx context.Context, tripID string) ([]domainlanguage.Phrase, error) {
	principal, ok := authentication.From(ctx)
	if !ok {
		return nil, nil
	}
	pack, err := p.service.Phrases(ctx, principal, languageusecase.PhraseRequest{
		TripID: tripID, Categories: []string{string(domainlanguage.Emergency), string(domainlanguage.General)},
	})
	if err != nil {
		return nil, err
	}
	return pack.Phrases, nil
}
