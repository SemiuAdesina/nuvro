package contract_test

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/apperror"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/alert"
	commercedomain "github.com/SemiuAdesina/nuvro/backend/internal/domain/commerce"
	feedbackdomain "github.com/SemiuAdesina/nuvro/backend/internal/domain/feedback"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/itinerary"
	personalizationdomain "github.com/SemiuAdesina/nuvro/backend/internal/domain/personalization"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/pricewatch"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/replan"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/retention"
	sessiondomain "github.com/SemiuAdesina/nuvro/backend/internal/domain/session"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/trip"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/admission"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/cursor"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/id"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/replay"
	nuvrohttp "github.com/SemiuAdesina/nuvro/backend/internal/transport/http"
	"github.com/SemiuAdesina/nuvro/backend/internal/transport/http/authentication"
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
	"github.com/SemiuAdesina/nuvro/backend/internal/usecase/plan"
	"github.com/SemiuAdesina/nuvro/backend/internal/usecase/profile"
)

const travellerID = "3f1c9c8e-0e3a-4f0a-9b1f-1a2b3c4d5e6f"
const otherTravellerID = "9a8b7c6d-5e4f-4a3b-8c2d-1e0f9a8b7c6d"
const credential = "synthetic-contract-credential"

type store struct {
	profiles        map[string]user.Profile
	preferences     map[string]user.Preferences
	trips           map[string]trip.Trip
	roles           map[string]trip.Role
	itinerary       itinerary.Itinerary
	hasItinerary    bool
	alerts          map[string]alert.Alert
	proposals       map[string]replan.Proposal
	handoffs        map[string]commercedomain.Handoff
	watches         map[string]pricewatch.Watch
	conversions     map[string]commercedomain.Conversion
	claimed         map[string]bool
	signals         []personalizationdomain.Signal
	learned         map[string]personalizationdomain.LearnedPreference
	feedback        map[string]feedbackdomain.TripFeedback
	historyLearning *bool
	deletions       map[string]retention.Deletion
	sessions        map[string]sessiondomain.Session
	sessionHashes   map[string]string
	closed          []string
}

func newStore() *store {
	return &store{
		profiles:    map[string]user.Profile{travellerID: user.DefaultProfile(travellerID)},
		preferences: map[string]user.Preferences{travellerID: user.DefaultPreferences()},
		trips:       map[string]trip.Trip{},
		roles:       map[string]trip.Role{},
	}
}

func (s *store) Within(ctx context.Context, operation func(context.Context) error) error {
	return operation(ctx)
}

func (s *store) Profile(_ context.Context, userID string) (user.Profile, error) {
	found, ok := s.profiles[userID]
	if !ok {
		return user.Profile{}, apperror.New(apperror.CodeNotFound, "Profile not found.")
	}
	return found, nil
}

func (s *store) SaveProfile(_ context.Context, value user.Profile) error {
	s.profiles[value.UserID] = value
	return nil
}

func (s *store) Preferences(_ context.Context, userID string) (user.Preferences, error) {
	found, ok := s.preferences[userID]
	if !ok {
		return user.Preferences{}, apperror.New(apperror.CodeNotFound, "Preferences not found.")
	}
	return found, nil
}

func (s *store) SavePreferences(_ context.Context, userID string, value user.Preferences) error {
	s.preferences[userID] = value
	return nil
}

func (s *store) Insert(_ context.Context, value trip.Trip) error {
	s.trips[value.ID] = value
	return nil
}

func (s *store) AddOwner(_ context.Context, tripID, userID string) error {
	s.roles[tripID+"|"+userID] = trip.Owner
	return nil
}

func (s *store) Save(_ context.Context, value trip.Trip) error {
	s.trips[value.ID] = value
	return nil
}

func (s *store) Get(_ context.Context, tripID string) (trip.Trip, error) {
	found, ok := s.trips[tripID]
	if !ok {
		return trip.Trip{}, apperror.New(apperror.CodeNotFound, "Trip not found.")
	}
	return found, nil
}

func (s *store) Role(_ context.Context, tripID, userID string) (trip.Role, error) {
	role, ok := s.roles[tripID+"|"+userID]
	if !ok {
		return "", apperror.New(apperror.CodeNotFound, "Trip not found.")
	}
	return role, nil
}

func (s *store) ListForMember(_ context.Context, userID string, after *cursor.Position, limit int32) ([]trip.Trip, error) {
	visible := make([]trip.Trip, 0, len(s.trips))
	for identifier, value := range s.trips {
		if s.roles[identifier+"|"+userID] != "" {
			visible = append(visible, value)
		}
	}
	slices.SortFunc(visible, func(left, right trip.Trip) int {
		if ordered := right.CreatedAt.Compare(left.CreatedAt); ordered != 0 {
			return ordered
		}
		return strings.Compare(right.ID, left.ID)
	})
	found := make([]trip.Trip, 0, limit)
	for _, value := range visible {
		if after != nil && !isBefore(value, *after) {
			continue
		}
		found = append(found, value)
		if int32(len(found)) == limit {
			break
		}
	}
	return found, nil
}

func isBefore(value trip.Trip, position cursor.Position) bool {
	if ordered := value.CreatedAt.Compare(position.CreatedAt); ordered != 0 {
		return ordered < 0
	}
	return value.ID < position.ID
}

type environment struct {
	handler http.Handler
	store   *store
	now     time.Time
	calls   int
}

type option func(*nuvrohttp.RouterDependencies)

func withLimiter(limiter admission.Limiter) option {
	return func(deps *nuvrohttp.RouterDependencies) { deps.Limiter = limiter }
}

func withReplays(store replay.Store) option {
	return func(deps *nuvrohttp.RouterDependencies) { deps.Replays = store }
}

func newEnvironment(t *testing.T, options ...option) *environment {
	t.Helper()
	repository := newStore()
	now := time.Date(2026, 9, 9, 12, 0, 0, 0, time.UTC)
	timer := clock.NewFixed(now)
	verifier, err := authentication.NewDevelopmentVerifier(true, credential, "contract-subject")
	if err != nil {
		t.Fatal(err)
	}
	codec, err := cursor.New([]byte(strings.Repeat("c", 32)))
	if err != nil {
		t.Fatal(err)
	}
	tokens := sessionService(t, repository, verifier, timer)
	sessions := authentication.NewSessions(verifier, resolver{}).WithTokens(tokens)
	dependencies := nuvrohttp.RouterDependencies{
		Logger: slog.New(slog.NewJSONHandler(&bytes.Buffer{}, nil)), Clock: timer, Identifiers: id.NewRandom(),
		Health:  health.NewHandler(timer),
		Profile: profilehandler.NewHandler(profile.NewService(repository, repository), eraser(t, repository, timer).WithSessions(tokens)),
		Trips: tripshandler.NewHandler(
			plan.NewCreator(repository, repository, id.NewUUID(), timer),
			plan.NewReader(repository),
			plan.NewLister(repository, codec),
			plan.NewUpdater(repository, repository, timer),
		),
		Discover:     discoverhandler.NewHandler(discoveryService(t, repository, timer)),
		Intelligence: intelligencehandler.NewHandler(destinationService(t, repository, timer)),
		Journey:      journeyhandler.NewHandler(journeyService(t, repository, timer)),
		Runtime: runtimehandler.NewHandler(travelService(t, repository, timer), alertService(t, repository, timer),
			languageService(t, repository), phraseSource{}),
		Commerce:        commercehandler.NewHandler(commerceService(t, repository, timer), timer),
		Personalization: personalizationhandler.NewHandler(personalizationService(t, repository, timer)),
		Authenticator:   sessions,
		Auth:            authhandler.NewHandler(tokens),
	}
	for _, apply := range options {
		apply(&dependencies)
	}
	return &environment{handler: nuvrohttp.NewRouter(dependencies), store: repository, now: now}
}

type resolver struct{}

func (resolver) Resolve(_ context.Context, identity user.Identity) (user.Principal, error) {
	return user.Principal{UserID: travellerID, Subject: identity.Subject, Provider: identity.Provider}, nil
}
