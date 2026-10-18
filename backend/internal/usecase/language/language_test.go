package language_test

import (
	"context"
	"testing"

	"github.com/SemiuAdesina/nuvro/backend/internal/apperror"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/destination"
	domain "github.com/SemiuAdesina/nuvro/backend/internal/domain/language"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/trip"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
	engine "github.com/SemiuAdesina/nuvro/backend/internal/engine/language"
	usecase "github.com/SemiuAdesina/nuvro/backend/internal/usecase/language"
)

const tripID = "8c2f21b0-4a55-4b39-9f0b-2c7f9c1a1a11"
const owner = "3f1c9c8e-0e3a-4f0a-9b1f-1a2b3c4d5e6f"

type trips struct{ destinationID *string }

func (t trips) Read(_ context.Context, principal user.Principal, _ string) (trip.Trip, error) {
	if principal.UserID == "" {
		return trip.Trip{}, apperror.New(apperror.CodeUnauthenticated, "Authentication is required.")
	}
	return trip.Trip{ID: tripID, OwnerID: owner, DestinationID: t.destinationID}, nil
}

type catalog struct{ country string }

func (c catalog) Destination(context.Context, string) (destination.Destination, error) {
	return destination.Destination{ID: "madeira-pt", CountryCode: c.country, Timezone: "Atlantic/Madeira"}, nil
}

type profiles struct{ language string }

func (p profiles) Profile(context.Context, string) (user.Profile, error) {
	return user.Profile{UserID: owner, PreferredLanguage: p.language}, nil
}

type translator struct {
	result     domain.Translation
	err        error
	pack       domain.Pack
	seenSource domain.Code
	seenTarget domain.Code
	seenText   string
}

func (t *translator) Translate(_ context.Context, request engine.Request) (domain.Translation, error) {
	t.seenText = request.Text
	source, _ := domain.Parse(request.Source)
	target, _ := domain.Parse(request.Target)
	t.seenSource = source
	t.seenTarget = target
	if t.err != nil {
		return domain.Translation{}, t.err
	}
	result := t.result
	result.Source = source
	result.Target = target
	result.SourceText = request.Text
	return result, nil
}

func (t *translator) Phrases(_ context.Context, source, target domain.Code, categories []domain.Category) domain.Pack {
	t.seenSource = source
	t.seenTarget = target
	pack := t.pack
	pack.Source = source
	pack.Target = target
	if len(pack.Phrases) == 0 {
		pack.Phrases = domain.EssentialPhrases(categories...)
	}
	return pack
}

func (t *translator) Configured() bool { return t.err == nil }

func service(t *testing.T, appLanguage string, engineTranslator *translator) *usecase.Service {
	t.Helper()
	destinationID := "madeira-pt"
	return usecase.NewService(trips{destinationID: &destinationID}, catalog{country: "PT"},
		profiles{language: appLanguage}, engineTranslator)
}

func principal() user.Principal { return user.Principal{UserID: owner} }

func TestDestinationDrivesTheDefaultTargetLanguage(t *testing.T) {
	source := &translator{result: domain.Translation{Translated: "Ola", Confidence: "high"}}
	if _, err := service(t, "en", source).Translate(t.Context(), principal(), usecase.TranslateRequest{
		TripID: tripID, Text: "Hello",
	}); err != nil {
		t.Fatal(err)
	}
	if source.seenTarget.String() != "pt-PT" || source.seenSource.String() != "en" {
		t.Fatalf("resolved %s to %s", source.seenSource, source.seenTarget)
	}
}

func TestAppLanguageIsRespectedNotAssumed(t *testing.T) {
	source := &translator{result: domain.Translation{Translated: "Bonjour", Confidence: "high"}}
	if _, err := service(t, "fr-FR", source).Translate(t.Context(), principal(), usecase.TranslateRequest{
		TripID: tripID, Text: "Hello",
	}); err != nil {
		t.Fatal(err)
	}
	if source.seenSource.String() != "fr-FR" {
		t.Fatalf("source = %s", source.seenSource)
	}
}

func TestExplicitLanguagesOverrideDefaults(t *testing.T) {
	source := &translator{result: domain.Translation{Translated: "Hola", Confidence: "high"}}
	if _, err := service(t, "en", source).Translate(t.Context(), principal(), usecase.TranslateRequest{
		TripID: tripID, Text: "Hello", Source: "en-GB", Target: "es-ES",
	}); err != nil {
		t.Fatal(err)
	}
	if source.seenSource.String() != "en-GB" || source.seenTarget.String() != "es-ES" {
		t.Fatalf("resolved %s to %s", source.seenSource, source.seenTarget)
	}
}

func TestUnknownLanguageTagsAreRejectedBeforeTheProvider(t *testing.T) {
	source := &translator{result: domain.Translation{Translated: "Ola"}}
	subject := service(t, "en", source)
	if _, err := subject.Translate(t.Context(), principal(), usecase.TranslateRequest{
		TripID: tripID, Text: "Hello", Target: "Portuguese",
	}); apperror.CodeOf(err) != apperror.CodeInvalidInput {
		t.Fatal("a display name was accepted")
	}
	if source.seenText != "" {
		t.Fatal("an invalid request reached the provider")
	}
}

func TestProviderFailureBecomesAnUnavailableError(t *testing.T) {
	source := &translator{err: engine.ErrUnavailable}
	if _, err := service(t, "en", source).Translate(t.Context(), principal(), usecase.TranslateRequest{
		TripID: tripID, Text: "Hello",
	}); apperror.CodeOf(err) != apperror.CodeUnavailable {
		t.Fatal("a provider outage was not reported as unavailable")
	}
	unsupported := &translator{err: engine.ErrUnsupported}
	if _, err := service(t, "en", unsupported).Translate(t.Context(), principal(), usecase.TranslateRequest{
		TripID: tripID, Text: "Hello",
	}); apperror.CodeOf(err) != apperror.CodeInvalidInput {
		t.Fatal("an unsupported pair was not reported as a client problem")
	}
	invalid := &translator{err: engine.ErrInvalidText}
	if _, err := service(t, "en", invalid).Translate(t.Context(), principal(), usecase.TranslateRequest{
		TripID: tripID, Text: "Hello",
	}); apperror.CodeOf(err) != apperror.CodeInvalidInput {
		t.Fatal("invalid text was not reported as a client problem")
	}
}

func TestPhraseCategoriesAreValidated(t *testing.T) {
	source := &translator{}
	subject := service(t, "en", source)
	pack, err := subject.Phrases(t.Context(), principal(), usecase.PhraseRequest{
		TripID: tripID, Categories: []string{"emergency"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(pack.Phrases) == 0 {
		t.Fatal("no phrases returned")
	}
	if _, err := subject.Phrases(t.Context(), principal(), usecase.PhraseRequest{
		TripID: tripID, Categories: []string{"smalltalk"},
	}); apperror.CodeOf(err) != apperror.CodeInvalidInput {
		t.Fatal("an invented category was accepted")
	}
}

func TestATripWithNoDestinationStillTranslates(t *testing.T) {
	source := &translator{result: domain.Translation{Translated: "Hello", Confidence: "unknown"}}
	subject := usecase.NewService(trips{}, catalog{country: "PT"}, profiles{language: "en"}, source)
	if _, err := subject.Translate(t.Context(), principal(), usecase.TranslateRequest{
		TripID: tripID, Text: "Hello", Target: "pt-PT",
	}); err != nil {
		t.Fatal(err)
	}
	if source.seenTarget.String() != "pt-PT" {
		t.Fatalf("target = %s", source.seenTarget)
	}
}

func TestAnonymousCallersAreRefused(t *testing.T) {
	source := &translator{result: domain.Translation{Translated: "Ola"}}
	subject := service(t, "en", source)
	if _, err := subject.Translate(t.Context(), user.Principal{}, usecase.TranslateRequest{
		TripID: tripID, Text: "Hello",
	}); err == nil {
		t.Fatal("an anonymous caller translated text")
	}
	if _, err := subject.Phrases(t.Context(), user.Principal{}, usecase.PhraseRequest{TripID: tripID}); err == nil {
		t.Fatal("an anonymous caller requested phrases")
	}
}
