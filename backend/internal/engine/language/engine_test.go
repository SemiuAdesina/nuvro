package language_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	domain "github.com/SemiuAdesina/nuvro/backend/internal/domain/language"
	engine "github.com/SemiuAdesina/nuvro/backend/internal/engine/language"
)

type provider struct {
	result    engine.Result
	err       error
	supported map[domain.Code]bool
	seen      []engine.Query
}

func (p *provider) Translate(_ context.Context, query engine.Query) (engine.Result, error) {
	p.seen = append(p.seen, query)
	return p.result, p.err
}

func (p *provider) Supports(code domain.Code) bool {
	if p.supported == nil {
		return true
	}
	return p.supported[code]
}

func working() *provider {
	return &provider{result: engine.Result{
		Source: domain.Code("en"), Translated: "Ola", Confidence: "high", Provider: "fixture",
	}}
}

func TestTranslationRequiresRealProviderOutput(t *testing.T) {
	subject := engine.New(working())
	result, err := subject.Translate(t.Context(), engine.Request{
		Source: "en", Target: "pt-PT", Text: "  Hello  ",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Translated != "Ola" || result.SourceText != "Hello" {
		t.Fatalf("result = %+v", result)
	}
	if result.Source.String() != "en" || result.Target.String() != "pt-PT" {
		t.Fatalf("language pair = %s to %s", result.Source, result.Target)
	}
	if result.Detected || result.Confidence != "high" {
		t.Fatalf("detection or confidence wrong: %+v", result)
	}
}

func TestSourceDetectionUsesTheProviderAnswer(t *testing.T) {
	source := working()
	source.result.Source = domain.Code("es")
	result, err := engine.New(source).Translate(t.Context(), engine.Request{Target: "pt-PT", Text: "Hola"})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Detected || result.Source.String() != "es" {
		t.Fatalf("detection = %+v", result)
	}
	silent := working()
	silent.result.Source = ""
	if _, err := engine.New(silent).Translate(t.Context(), engine.Request{Target: "pt-PT", Text: "Hola"}); err == nil {
		t.Fatal("a provider that detected nothing was treated as successful")
	}
}

func TestUnconfiguredAndFailingProvidersAreHonest(t *testing.T) {
	unconfigured := engine.New(nil)
	if unconfigured.Configured() {
		t.Fatal("an unconfigured engine claimed to be configured")
	}
	if _, err := unconfigured.Translate(t.Context(), engine.Request{Target: "pt-PT", Text: "Hello"}); !errors.Is(err, engine.ErrUnavailable) {
		t.Fatalf("error = %v", err)
	}
	failing := engine.New(&provider{err: errors.New("provider down")})
	if _, err := failing.Translate(t.Context(), engine.Request{Target: "pt-PT", Text: "Hello"}); !errors.Is(err, engine.ErrUnavailable) {
		t.Fatalf("error = %v", err)
	}
	empty := working()
	empty.result.Translated = "   "
	if _, err := engine.New(empty).Translate(t.Context(), engine.Request{Target: "pt-PT", Text: "Hello"}); err == nil {
		t.Fatal("an empty translation was accepted")
	}
}

func TestUnsupportedLanguagesAreRefused(t *testing.T) {
	limited := working()
	limited.supported = map[domain.Code]bool{domain.Code("pt-PT"): true}
	if _, err := engine.New(limited).Translate(t.Context(), engine.Request{
		Source: "sw-TZ", Target: "pt-PT", Text: "Hello",
	}); !errors.Is(err, engine.ErrUnsupported) {
		t.Fatalf("error = %v", err)
	}
	if _, err := engine.New(working()).Translate(t.Context(), engine.Request{
		Target: "Portuguese", Text: "Hello",
	}); !errors.Is(err, engine.ErrUnsupported) {
		t.Fatal("a display name was accepted as a language tag")
	}
}

func TestEmptyAndOversizedTextIsRefusedBeforeTheProvider(t *testing.T) {
	source := working()
	subject := engine.New(source)
	for name, text := range map[string]string{
		"empty":     "",
		"blank":     "    ",
		"oversized": strings.Repeat("a", engine.MaximumTextBytes+1),
	} {
		if _, err := subject.Translate(t.Context(), engine.Request{Target: "pt-PT", Text: text}); !errors.Is(err, engine.ErrInvalidText) {
			t.Fatalf("%s produced %v", name, err)
		}
	}
	if len(source.seen) != 0 {
		t.Fatal("invalid text reached the provider")
	}
}

func TestContextIsBoundedBeforeLeavingTheEngine(t *testing.T) {
	source := working()
	if _, err := engine.New(source).Translate(t.Context(), engine.Request{
		Target: "pt-PT", Text: "Hello", Context: strings.Repeat("c", 900),
	}); err != nil {
		t.Fatal(err)
	}
	if len(source.seen[0].Context) > 200 {
		t.Fatalf("context of %d bytes was forwarded", len(source.seen[0].Context))
	}
}

func TestLanguageResolutionPrefersRequestThenDestination(t *testing.T) {
	source, target := engine.Resolve(engine.Context{
		AppLanguage: domain.Code("en"), DestinationRegion: "PT",
	})
	if source.String() != "en" || target.String() != "pt-PT" {
		t.Fatalf("resolved %s to %s", source, target)
	}
	_, explicit := engine.Resolve(engine.Context{
		AppLanguage: domain.Code("en"), DestinationRegion: "PT",
		Target: domain.Code("fr-FR"), HasTarget: true,
	})
	if explicit.String() != "fr-FR" {
		t.Fatalf("an explicit target was overridden: %s", explicit)
	}
	frenchApp, frenchTarget := engine.Resolve(engine.Context{
		AppLanguage: domain.Code("fr-FR"), DestinationRegion: "PT",
	})
	if frenchApp.String() != "fr-FR" || frenchTarget.String() != "pt-PT" {
		t.Fatalf("a different app language broke resolution: %s to %s", frenchApp, frenchTarget)
	}
	unknownRegion, unknownTarget := engine.Resolve(engine.Context{
		AppLanguage: domain.Code("en"), DestinationRegion: "ZZ",
	})
	if unknownRegion != unknownTarget {
		t.Fatal("an unknown region invented a target language")
	}
}

func TestPhrasePacksAreCategorisedAndTranslatedWhenPossible(t *testing.T) {
	source := working()
	pack := engine.New(source).Phrases(t.Context(), domain.Code("en"), domain.Code("pt-PT"),
		[]domain.Category{domain.Emergency})
	if len(pack.Phrases) == 0 {
		t.Fatal("no emergency phrases")
	}
	for _, phrase := range pack.Phrases {
		if phrase.Category != domain.Emergency {
			t.Fatalf("category = %q", phrase.Category)
		}
		if !phrase.HasTargetText || phrase.TargetText == "" {
			t.Fatalf("phrase was not translated: %+v", phrase)
		}
	}
	unconfigured := engine.New(nil).Phrases(t.Context(), domain.Code("en"), domain.Code("pt-PT"), nil)
	if len(unconfigured.Phrases) == 0 {
		t.Fatal("no phrases without a provider")
	}
	for _, phrase := range unconfigured.Phrases {
		if phrase.HasTargetText {
			t.Fatal("a translation was invented without a provider")
		}
	}
}

func TestEmergencyPhrasesStayLiteralAndSafe(t *testing.T) {
	phrases := domain.EssentialPhrases(domain.Emergency, domain.Medical)
	if len(phrases) < 4 {
		t.Fatalf("only %d emergency phrases", len(phrases))
	}
	for _, phrase := range phrases {
		lowered := strings.ToLower(phrase.SourceText)
		for _, unsafe := range []string{"take ", "dose", "mg", "diagnos", "treat"} {
			if strings.Contains(lowered, unsafe) {
				t.Fatalf("phrase %q reads as medical advice", phrase.SourceText)
			}
		}
	}
}
