package http

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"

	"github.com/SemiuAdesina/nuvro/backend/internal/platform/admission"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clientip"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/id"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/replay"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/telemetry"
	"github.com/SemiuAdesina/nuvro/backend/internal/transport/http/api"
	"github.com/SemiuAdesina/nuvro/backend/internal/transport/http/authentication"
	authhandler "github.com/SemiuAdesina/nuvro/backend/internal/transport/http/handler/auth"
	commercehandler "github.com/SemiuAdesina/nuvro/backend/internal/transport/http/handler/commerce"
	discoverhandler "github.com/SemiuAdesina/nuvro/backend/internal/transport/http/handler/discover"
	"github.com/SemiuAdesina/nuvro/backend/internal/transport/http/handler/health"
	intelligencehandler "github.com/SemiuAdesina/nuvro/backend/internal/transport/http/handler/intelligence"
	journeyhandler "github.com/SemiuAdesina/nuvro/backend/internal/transport/http/handler/journey"
	personalizationhandler "github.com/SemiuAdesina/nuvro/backend/internal/transport/http/handler/personalization"
	"github.com/SemiuAdesina/nuvro/backend/internal/transport/http/handler/profile"
	runtimehandler "github.com/SemiuAdesina/nuvro/backend/internal/transport/http/handler/runtime"
	"github.com/SemiuAdesina/nuvro/backend/internal/transport/http/handler/trips"
	"github.com/SemiuAdesina/nuvro/backend/internal/transport/http/httperror"
	"github.com/SemiuAdesina/nuvro/backend/internal/transport/http/middleware"
)

type RouterDependencies struct {
	Logger          *slog.Logger
	Clock           clock.Clock
	Identifiers     id.Generator
	Health          *health.Handler
	Profile         *profile.Handler
	Trips           *trips.Handler
	Discover        *discoverhandler.Handler
	Intelligence    *intelligencehandler.Handler
	Journey         *journeyhandler.Handler
	Runtime         *runtimehandler.Handler
	Commerce        *commercehandler.Handler
	Personalization *personalizationhandler.Handler
	Auth            *authhandler.Handler
	Authenticator   authentication.Authenticator
	Tracer          trace.Tracer
	Metrics         *telemetry.Metrics
	Limiter         admission.Limiter
	Replays         replay.Store
	Addresses       *clientip.Resolver
}

func NewRouter(deps RouterDependencies) http.Handler {
	router := chi.NewRouter()

	router.Use(middleware.RequestID(deps.Identifiers))
	if deps.Tracer == nil {
		deps.Tracer = noop.NewTracerProvider().Tracer("nuvro")
	}
	router.Use(middleware.Trace(deps.Tracer, deps.Metrics))
	router.Use(middleware.AccessLog(deps.Logger, deps.Clock))
	router.Use(middleware.Recover(deps.Logger))
	router.Use(middleware.SecurityHeaders)
	router.Use(middleware.BodyLimit(64 << 10))

	router.NotFound(httperror.NotFoundHandler())
	router.MethodNotAllowed(httperror.MethodNotAllowedHandler())

	document, err := api.GetSpec()
	if err != nil {
		panic("generated API specification could not load")
	}
	if err := document.Validate(context.Background()); err != nil {
		panic("generated API specification is invalid")
	}
	handlers := server{systemOperations: deps.Health, profileOperations: deps.Profile, tripOperations: deps.Trips, discoverOperations: deps.Discover, intelligenceOperations: deps.Intelligence, journeyOperations: deps.Journey, runtimeOperations: deps.Runtime, commerceOperations: deps.Commerce, personalizationOperations: deps.Personalization, authOperations: deps.Auth}
	strict := api.NewStrictHandlerWithOptions(handlers, nil, api.StrictHTTPServerOptions{
		RequestErrorHandlerFunc: func(w http.ResponseWriter, _ *http.Request, _ error) {
			httperror.WriteStatus(w, 400, "validation_error", "Request is invalid.")
		},
		ResponseErrorHandlerFunc: func(w http.ResponseWriter, _ *http.Request, err error) { httperror.Write(w, err) },
	})
	api.HandlerWithOptions(strict, api.ChiServerOptions{
		BaseRouter: router,
		ErrorHandlerFunc: func(w http.ResponseWriter, _ *http.Request, _ error) {
			httperror.WriteStatus(w, http.StatusBadRequest, "validation_error", "Request is invalid.")
		},
		Middlewares: []api.MiddlewareFunc{
			middleware.Idempotent(deps.Replays, principalSubject),
			middleware.Admit(deps.Limiter, principalSubject),
			middleware.Validate(document, authentication.Validator(deps.Authenticator)),
			middleware.Guard(deps.Limiter, deps.Addresses),
		},
	})

	return router
}

func principalSubject(r *http.Request) (string, bool) {
	principal, found := authentication.From(r.Context())
	if !found || principal.UserID == "" {
		return "", false
	}
	return principal.UserID, true
}
