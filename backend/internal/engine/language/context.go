package language

import (
	"context"

	domain "github.com/SemiuAdesina/nuvro/backend/internal/domain/language"
)

type Context struct {
	AppLanguage       domain.Code
	DestinationRegion string
	Spoken            []domain.Code
	Target            domain.Code
	HasTarget         bool
}

func Resolve(request Context) (source domain.Code, target domain.Code) {
	source = request.AppLanguage
	if !source.Valid() {
		source = domain.Code("en")
	}
	if request.HasTarget && request.Target.Valid() {
		return source, request.Target
	}
	if regional, found := domain.ForRegion(request.DestinationRegion); found {
		if regional.SameLanguage(source) {
			return source, regional
		}
		return source, regional
	}
	return source, source
}

func (e *Engine) Phrases(ctx context.Context, source, target domain.Code, categories []domain.Category) domain.Pack {
	pack := domain.Pack{Source: source, Target: target, Phrases: domain.EssentialPhrases(categories...)}
	if e.provider == nil || source.SameLanguage(target) {
		return pack
	}
	for index := range pack.Phrases {
		result, err := e.provider.Translate(ctx, Query{
			Source: source, Target: target, Text: pack.Phrases[index].SourceText, Context: "travel_phrase",
		})
		if err != nil || result.Translated == "" {
			continue
		}
		pack.Phrases[index].TargetText = result.Translated
		pack.Phrases[index].HasTargetText = true
	}
	return pack
}
