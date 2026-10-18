package discover_test

import (
	"context"
	"testing"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/personalization"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
)

type learning struct {
	context personalization.Context
	err     error
	calls   int
}

func (l *learning) Context(context.Context, user.Principal) (personalization.Context, error) {
	l.calls++
	return l.context, l.err
}

func learned(keys ...personalization.Key) *learning {
	value := personalization.EmptyContext(traveller, true, evaluatedAt)
	for _, key := range keys {
		value.Learned = append(value.Learned, personalization.Preference{
			Key: key, Origin: personalization.InferredLearned, Confidence: personalization.High,
			Weight: personalization.WeightFor(personalization.InferredLearned, personalization.High),
		})
	}
	return &learning{context: value}
}

func TestLearnedInterestsWidenTheSearchWithoutChangingConstraints(t *testing.T) {
	baseline := newService(t, profileFor("NG", "LOS", "NGN"), &catalog{destinations: destinations()}, nigeriaProviders())
	personalised := newService(t, profileFor("NG", "LOS", "NGN"), &catalog{destinations: destinations()}, nigeriaProviders()).
		WithPersonalization(learned(personalization.InterestCoastal, personalization.InterestFood))
	request := warmBeachRequest(t, 5_000_000_00, "NGN")
	request.Overrides.Interests = nil
	plain, err := baseline.Matches(t.Context(), user.Principal{UserID: traveller}, withPassports(request, "NG"))
	if err != nil {
		t.Fatal(err)
	}
	tuned, err := personalised.Matches(t.Context(), user.Principal{UserID: traveller}, withPassports(request, "NG"))
	if err != nil {
		t.Fatal(err)
	}
	if len(tuned.Intent.Preferences.Interests) <= len(plain.Intent.Preferences.Interests) {
		t.Fatalf("learning did not add interests: %v then %v",
			plain.Intent.Preferences.Interests, tuned.Intent.Preferences.Interests)
	}
	if tuned.Intent.Budget.Minor() != plain.Intent.Budget.Minor() {
		t.Fatal("personalization changed the traveller's budget")
	}
	if tuned.Intent.Origin.DepartureAirport != plain.Intent.Origin.DepartureAirport {
		t.Fatal("personalization changed the departure airport")
	}
	if len(tuned.Intent.Traveler.PassportCountries) != len(plain.Intent.Traveler.PassportCountries) {
		t.Fatal("personalization changed travel documents")
	}
	for _, recommendation := range tuned.Result.Recommendations {
		for _, original := range plain.Result.Recommendations {
			if recommendation.Destination.ID != original.Destination.ID {
				continue
			}
			if recommendation.Feasibility.Access.Status != original.Feasibility.Access.Status {
				t.Fatal("personalization changed a feasibility assessment")
			}
		}
	}
}

func TestExplicitRequestPreferencesSurvivePersonalization(t *testing.T) {
	service := newService(t, profileFor("NG", "LOS", "NGN"), &catalog{destinations: destinations()}, nigeriaProviders()).
		WithPersonalization(learned(personalization.PacePacked, personalization.StayCentralArea))
	request := warmBeachRequest(t, 5_000_000_00, "NGN")
	request.Overrides.Pace = "relaxed"
	request.Overrides.Crowds = "quiet"
	result, err := service.Matches(t.Context(), user.Principal{UserID: traveller}, withPassports(request, "NG"))
	if err != nil {
		t.Fatal(err)
	}
	if result.Intent.Preferences.Pace != "relaxed" {
		t.Fatalf("pace = %q, want the traveller's own choice", result.Intent.Preferences.Pace)
	}
	if result.Intent.Preferences.Crowds != "quiet" {
		t.Fatalf("crowds = %q, want the traveller's own choice", result.Intent.Preferences.Crowds)
	}
}

func TestExplicitProfilePaceBeatsAnInference(t *testing.T) {
	service := newService(t, profileFor("NG", "LOS", "NGN"), &catalog{destinations: destinations()}, nigeriaProviders()).
		WithPersonalization(learned(personalization.PaceRelaxed))
	request := warmBeachRequest(t, 5_000_000_00, "NGN")
	request.Overrides.Pace = ""
	result, err := service.Matches(t.Context(), user.Principal{UserID: traveller}, withPassports(request, "NG"))
	if err != nil {
		t.Fatal(err)
	}
	if result.Intent.Preferences.Pace != "balanced" {
		t.Fatalf("pace = %q, want the profile value the traveller set", result.Intent.Preferences.Pace)
	}
}

func TestLearningFillsAPreferenceTheTravellerNeverSet(t *testing.T) {
	unset := profileFor("NG", "LOS", "NGN")
	unset.preferences.Pace = ""
	service := newService(t, unset, &catalog{destinations: destinations()}, nigeriaProviders()).
		WithPersonalization(learned(personalization.PaceRelaxed))
	request := warmBeachRequest(t, 5_000_000_00, "NGN")
	request.Overrides.Pace = ""
	result, err := service.Matches(t.Context(), user.Principal{UserID: traveller}, withPassports(request, "NG"))
	if err != nil {
		t.Fatal(err)
	}
	if result.Intent.Preferences.Pace != "relaxed" {
		t.Fatalf("pace = %q, want the learned value", result.Intent.Preferences.Pace)
	}
}

func TestDisabledLearningLeavesTheRequestUntouched(t *testing.T) {
	disabled := learned(personalization.InterestCoastal, personalization.PaceRelaxed)
	disabled.context.HistoryLearningEnabled = false
	baseline := newService(t, profileFor("NG", "LOS", "NGN"), &catalog{destinations: destinations()}, nigeriaProviders())
	personalised := newService(t, profileFor("NG", "LOS", "NGN"), &catalog{destinations: destinations()}, nigeriaProviders()).
		WithPersonalization(disabled)
	request := warmBeachRequest(t, 5_000_000_00, "NGN")
	request.Overrides.Interests = nil
	request.Overrides.Pace = ""
	plain, err := baseline.Matches(t.Context(), user.Principal{UserID: traveller}, withPassports(request, "NG"))
	if err != nil {
		t.Fatal(err)
	}
	tuned, err := personalised.Matches(t.Context(), user.Principal{UserID: traveller}, withPassports(request, "NG"))
	if err != nil {
		t.Fatal(err)
	}
	if len(tuned.Intent.Preferences.Interests) != len(plain.Intent.Preferences.Interests) {
		t.Fatal("disabled learning still widened the search")
	}
	if tuned.Intent.Preferences.Pace != plain.Intent.Preferences.Pace {
		t.Fatal("disabled learning still set a pace")
	}
	if ordered(tuned) != ordered(plain) {
		t.Fatal("disabled learning changed the ranking")
	}
}

func TestPersonalizationFailureDoesNotBreakDiscovery(t *testing.T) {
	failing := &learning{err: errPersonalization{}}
	service := newService(t, profileFor("NG", "LOS", "NGN"), &catalog{destinations: destinations()}, nigeriaProviders()).
		WithPersonalization(failing)
	result, err := service.Matches(t.Context(), user.Principal{UserID: traveller},
		withPassports(warmBeachRequest(t, 5_000_000_00, "NGN"), "NG"))
	if err != nil {
		t.Fatalf("a personalization outage failed discovery: %v", err)
	}
	if len(result.Result.Recommendations) == 0 {
		t.Fatal("no recommendations after a personalization outage")
	}
	if failing.calls == 0 {
		t.Fatal("personalization was never consulted")
	}
}

func TestColdStartMatchesTheUnpersonalizedResult(t *testing.T) {
	empty := &learning{context: personalization.EmptyContext(traveller, true, evaluatedAt)}
	baseline := newService(t, profileFor("NG", "LOS", "NGN"), &catalog{destinations: destinations()}, nigeriaProviders())
	personalised := newService(t, profileFor("NG", "LOS", "NGN"), &catalog{destinations: destinations()}, nigeriaProviders()).
		WithPersonalization(empty)
	request := withPassports(warmBeachRequest(t, 5_000_000_00, "NGN"), "NG")
	plain, err := baseline.Matches(t.Context(), user.Principal{UserID: traveller}, request)
	if err != nil {
		t.Fatal(err)
	}
	tuned, err := personalised.Matches(t.Context(), user.Principal{UserID: traveller}, request)
	if err != nil {
		t.Fatal(err)
	}
	if ordered(tuned) != ordered(plain) {
		t.Fatalf("a new user got a different ranking: %v then %v", ordered(plain), ordered(tuned))
	}
}

type errPersonalization struct{}

func (errPersonalization) Error() string { return "personalization unavailable" }
