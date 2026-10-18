package app

import (
	"context"
	"io"
	"log/slog"
	"net/http"

	"github.com/SemiuAdesina/nuvro/backend/internal/adapter/external/httpclient"
	"github.com/SemiuAdesina/nuvro/backend/internal/adapter/postgres"
	redisadapter "github.com/SemiuAdesina/nuvro/backend/internal/adapter/redis"
	"github.com/SemiuAdesina/nuvro/backend/internal/transport/http/authentication"

	"github.com/SemiuAdesina/nuvro/backend/internal/engine/alerting"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clientip"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/config"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/id"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/logging"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/telemetry"
	nuvrohttp "github.com/SemiuAdesina/nuvro/backend/internal/transport/http"
	authhandler "github.com/SemiuAdesina/nuvro/backend/internal/transport/http/handler/auth"
	commercehandler "github.com/SemiuAdesina/nuvro/backend/internal/transport/http/handler/commerce"
	discoverhandler "github.com/SemiuAdesina/nuvro/backend/internal/transport/http/handler/discover"
	"github.com/SemiuAdesina/nuvro/backend/internal/transport/http/handler/health"
	intelligencehandler "github.com/SemiuAdesina/nuvro/backend/internal/transport/http/handler/intelligence"
	journeyhandler "github.com/SemiuAdesina/nuvro/backend/internal/transport/http/handler/journey"
	personalizationhandler "github.com/SemiuAdesina/nuvro/backend/internal/transport/http/handler/personalization"
	profilehandler "github.com/SemiuAdesina/nuvro/backend/internal/transport/http/handler/profile"
	runtimehandler "github.com/SemiuAdesina/nuvro/backend/internal/transport/http/handler/runtime"
	tripshandler "github.com/SemiuAdesina/nuvro/backend/internal/transport/http/handler/trips"
	"github.com/SemiuAdesina/nuvro/backend/internal/usecase/account"
	alertsusecase "github.com/SemiuAdesina/nuvro/backend/internal/usecase/alerts"
	destinationusecase "github.com/SemiuAdesina/nuvro/backend/internal/usecase/destination"
	"github.com/SemiuAdesina/nuvro/backend/internal/usecase/discover"
	journeyusecase "github.com/SemiuAdesina/nuvro/backend/internal/usecase/journey"
	"github.com/SemiuAdesina/nuvro/backend/internal/usecase/plan"
	"github.com/SemiuAdesina/nuvro/backend/internal/usecase/profile"
	"github.com/SemiuAdesina/nuvro/backend/internal/worker/jobs"
	"github.com/SemiuAdesina/nuvro/backend/internal/worker/runner"
)

func NewAPI(cfg config.Config, logDestination io.Writer) (*API, error) {
	logger, err := logging.New(logDestination, cfg)
	if err != nil {
		return nil, err
	}

	dependencies, err := openDependencies(cfg)
	if err != nil {
		return nil, err
	}
	metrics, err := telemetry.NewMetrics(dependencies.telemetry.Meter())
	if err != nil {
		_ = dependencies.close(context.Background())
		return nil, err
	}
	addresses, err := cfg.Proxy.Resolver()
	if err != nil {
		_ = dependencies.close(context.Background())
		return nil, err
	}
	timer := clock.NewSystem()
	product, err := newProductHandlers(cfg, dependencies.store, redisadapter.NewCache(dependencies.redis), timer, metrics)
	if err != nil {
		_ = dependencies.close(context.Background())
		return nil, err
	}
	admitted, err := newGates(cfg, dependencies.redis, timer)
	if err != nil {
		_ = dependencies.close(context.Background())
		return nil, err
	}
	router := newRouter(logger, dependencies, metrics, timer, addresses, product, admitted)

	server := &http.Server{
		Addr:              addressFromPort(cfg.HTTP.Port),
		Handler:           router,
		ReadHeaderTimeout: cfg.HTTP.ReadHeaderTimeout,
		ReadTimeout:       cfg.HTTP.ReadTimeout,
		WriteTimeout:      cfg.HTTP.WriteTimeout,
		IdleTimeout:       cfg.HTTP.IdleTimeout,
		MaxHeaderBytes:    cfg.HTTP.MaxHeaderBytes,
	}

	return &API{logger: logger, server: server, shutdownTimeout: cfg.HTTP.ShutdownTimeout, dependencies: dependencies}, nil
}

func newRouter(logger *slog.Logger, dependencies *dependencies, metrics *telemetry.Metrics,
	timer clock.Clock, addresses *clientip.Resolver, product productHandlers, admitted gates) http.Handler {
	return nuvrohttp.NewRouter(nuvrohttp.RouterDependencies{
		Logger:          logger,
		Tracer:          dependencies.telemetry.Tracer(),
		Metrics:         metrics,
		Clock:           timer,
		Identifiers:     id.NewRandom(),
		Addresses:       addresses,
		Health:          health.NewHandler(timer, dependencies.probes...),
		Profile:         product.profile,
		Intelligence:    product.intelligence,
		Journey:         product.journey,
		Runtime:         product.runtime,
		Commerce:        product.commerce,
		Personalization: product.personalization,
		Auth:            product.auth,
		Trips:           product.trips,
		Discover:        product.discover,
		Authenticator:   product.authenticator,
		Limiter:         admitted.limiter,
		Replays:         admitted.replays,
	})
}

func NewWorker(cfg config.Config, logDestination io.Writer) (*Worker, error) {
	logger, err := logging.New(logDestination, cfg)
	if err != nil {
		return nil, err
	}
	dependencies, err := openDependencies(cfg)
	if err != nil {
		return nil, err
	}
	timer := clock.NewSystem()
	jobs := workerJobs(cfg, dependencies, logger, timer)
	work := runner.New(logger, jobs...).WithLeases(leases(dependencies))
	return &Worker{logger: logger, runner: work, dependencies: dependencies}, nil
}

func leases(dependencies *dependencies) runner.Leases {
	if dependencies.store == nil || dependencies.postgres == nil {
		return nil
	}
	return func(name string) runner.Lease {
		return postgres.NewLease(dependencies.postgres.Pool, "nuvro:job:"+name)
	}
}

func workerJobs(cfg config.Config, dependencies *dependencies, logger *slog.Logger, timer clock.Clock) []runner.Job {
	if dependencies.store == nil {
		return nil
	}
	brains := newIntelligence(cfg, redisadapter.NewCache(dependencies.redis), destinationQueries{store: dependencies.store}, timer, nil)
	recorder := alertsusecase.NewService(nil, dependencies.store, id.NewUUID(), timer)
	return []runner.Job{
		jobs.NewAlertEvaluation(
			dependencies.store, newSnapshotEvaluator(dependencies.store, brains.weather, timer),
			alerting.New(), recorder, logger, timer, cfg.Runtime.AlertInterval,
		),
		jobs.NewRetentionSweep(dependencies.store, dependencies.store, revoker(cfg, dependencies.store, timer),
			logger, timer, cfg.Runtime.RetentionInterval),
		jobs.NewPriceWatchEvaluation(
			dependencies.store, nil, recorder, watchScheduler{}, id.NewUUID(),
			logger, timer, cfg.Runtime.PriceWatchInterval,
		),
	}
}

type productHandlers struct {
	profile         *profilehandler.Handler
	trips           *tripshandler.Handler
	discover        *discoverhandler.Handler
	intelligence    *intelligencehandler.Handler
	journey         *journeyhandler.Handler
	runtime         *runtimehandler.Handler
	commerce        *commercehandler.Handler
	personalization *personalizationhandler.Handler
	auth            *authhandler.Handler
	authenticator   authentication.Authenticator
}

func newProductHandlers(cfg config.Config, store *postgres.Store, cache *redisadapter.Cache,
	timer clock.Clock, metrics httpclient.Recorder) (productHandlers, error) {
	linker, err := newLinker(cfg, store, timer)
	if err != nil {
		return productHandlers{}, err
	}
	who, err := newIdentity(cfg, store, linker, timer)
	if err != nil {
		return productHandlers{}, err
	}
	if who.authenticator == nil || store == nil {
		return productHandlers{}, nil
	}
	cursors, err := newCursors(cfg)
	if err != nil {
		return productHandlers{}, err
	}
	engines, err := newDiscoveryEngines(cfg, timer)
	if err != nil {
		return productHandlers{}, err
	}
	brains := newIntelligence(cfg, cache, destinationQueries{store: store}, timer, metrics)
	tripReader := plan.NewReader(store)
	access := tripAccess{reader: tripReader, store: store}
	learning, err := newPersonalizationService(store, cache, access, timer)
	if err != nil {
		return productHandlers{}, err
	}
	eraser := account.NewEraser(store, id.NewUUID(), timer).WithLinker(linker).WithSessions(who.sessions)
	services := newRuntimeServices(store, brains, access, timer)
	return productHandlers{
		runtime:         runtimehandler.NewHandler(services.travel, services.alerts, services.language, services.phrases),
		commerce:        commercehandler.NewHandler(newCommerceService(cfg, store, cache, access, timer), timer),
		discover:        discoverhandler.NewHandler(discover.NewService(store, store, engines, timer).WithPersonalization(learning)),
		personalization: personalizationhandler.NewHandler(learning),
		intelligence:    intelligencehandler.NewHandler(destinationusecase.NewService(store, store, brains.describer, timer)),
		journey: journeyhandler.NewHandler(journeyusecase.NewService(journeyusecase.Dependencies{
			Trips: access, Catalog: store, Store: store,
			Preferences: store, Planner: brains.planner, Forecaster: brains.weather,
			Accountant: newAccountant(timer), Finder: brains.places, Identifiers: id.NewUUID(),
		}, timer)),
		profile: profilehandler.NewHandler(profile.NewService(store, store), eraser).WithLinker(linker),
		trips: tripshandler.NewHandler(
			plan.NewCreator(store, store, id.NewUUID(), timer),
			tripReader,
			plan.NewLister(store, cursors),
			plan.NewUpdater(store, store, timer),
		),
		authenticator: who.authenticator,
		auth:          authhandler.NewHandler(who.sessions),
	}, nil
}
