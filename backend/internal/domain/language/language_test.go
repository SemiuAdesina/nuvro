package language_test

import (
	"strings"
	"testing"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/language"
)

func TestTagParsingNormalisesCase(t *testing.T) {
	for raw, want := range map[string]string{
		"en": "en", "EN": "en", "pt-pt": "pt-PT", "PT-BR": "pt-BR",
		" fr-FR ": "fr-FR", "zh-hans-cn": "zh-Hans-CN",
	} {
		code, err := language.Parse(raw)
		if err != nil {
			t.Fatalf("%q was rejected: %v", raw, err)
		}
		if code.String() != want {
			t.Fatalf("%q became %q, want %q", raw, code, want)
		}
	}
}

func TestDisplayNamesAndJunkAreRejected(t *testing.T) {
	for _, raw := range []string{"English", "Portuguese", "", "   ", "e", "toolongtag", "en_US", "123"} {
		if _, err := language.Parse(raw); err == nil {
			t.Fatalf("%q was accepted as a language tag", raw)
		}
	}
}

func TestBaseAndComparison(t *testing.T) {
	portuguese, err := language.Parse("pt-PT")
	if err != nil {
		t.Fatal(err)
	}
	brazilian, err := language.Parse("pt-BR")
	if err != nil {
		t.Fatal(err)
	}
	if portuguese.Base().String() != "pt" || !portuguese.SameLanguage(brazilian) {
		t.Fatal("regional variants were not recognised as the same language")
	}
	english, err := language.Parse("en")
	if err != nil {
		t.Fatal(err)
	}
	if portuguese.SameLanguage(english) {
		t.Fatal("different languages compared equal")
	}
}

func TestRegionMappingIsExplicitNotGuessed(t *testing.T) {
	if code, found := language.ForRegion("pt"); !found || code.String() != "pt-PT" {
		t.Fatalf("PT resolved to %q %v", code, found)
	}
	if _, found := language.ForRegion("ZZ"); found {
		t.Fatal("an unknown region produced a language")
	}
	if code, found := language.ForRegion("NG"); !found || code.String() != "en-NG" {
		t.Fatalf("NG resolved to %q", code)
	}
}

func TestPhraseCatalogCoversEveryCategory(t *testing.T) {
	all := language.EssentialPhrases()
	seen := map[language.Category]int{}
	for _, phrase := range all {
		if phrase.Key == "" || phrase.SourceText == "" || !phrase.Category.Valid() {
			t.Fatalf("malformed phrase: %+v", phrase)
		}
		seen[phrase.Category]++
	}
	for _, category := range language.Categories() {
		if seen[category] == 0 {
			t.Fatalf("category %q has no phrases", category)
		}
	}
	if len(all) > 60 {
		t.Fatalf("the catalog has grown to %d phrases", len(all))
	}
	filtered := language.EssentialPhrases(language.Hotel)
	for _, phrase := range filtered {
		if phrase.Category != language.Hotel {
			t.Fatalf("filter leaked %q", phrase.Category)
		}
	}
}

func TestPackFiltersByCategory(t *testing.T) {
	pack := language.Pack{Phrases: language.EssentialPhrases()}
	emergency := pack.ByCategory(language.Emergency)
	if len(emergency) == 0 {
		t.Fatal("no emergency phrases in the pack")
	}
	for _, phrase := range emergency {
		if phrase.Category != language.Emergency {
			t.Fatal("pack filter leaked another category")
		}
	}
}

func FuzzParseLanguage(f *testing.F) {
	f.Add("en")
	f.Add("pt-PT")
	f.Add("English")
	f.Add(strings.Repeat("x", 40))
	f.Fuzz(func(t *testing.T, raw string) {
		code, err := language.Parse(raw)
		if err != nil {
			return
		}
		if !code.Valid() {
			t.Fatalf("%q produced an invalid code %q", raw, code)
		}
		if code.Base() == "" || strings.Contains(code.String(), " ") {
			t.Fatalf("%q produced a malformed code %q", raw, code)
		}
		if !code.SameLanguage(code) {
			t.Fatalf("%q is not the same language as itself", code)
		}
	})
}
