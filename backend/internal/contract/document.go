package contract

import (
	"context"
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
)

func Load(ctx context.Context, path string) (*openapi3.T, error) {
	loader := openapi3.NewLoader()
	loader.Context = ctx
	loader.IsExternalRefsAllowed = true
	document, err := loader.LoadFromFile(path)
	if err != nil {
		return nil, fmt.Errorf("contract: load: %w", err)
	}
	if document.OpenAPI != "3.1.0" {
		return nil, fmt.Errorf("contract: require OpenAPI 3.1.0")
	}
	if err := document.Validate(ctx); err != nil {
		return nil, fmt.Errorf("contract: validate: %w", err)
	}
	for _, item := range document.Paths.Map() {
		for _, operation := range item.Operations() {
			if operation.OperationID == "" || operation.Responses == nil || operation.Responses.Len() == 0 {
				return nil, fmt.Errorf("contract: operation requires operationId and responses")
			}
		}
	}
	return document, nil
}

func Bundle(ctx context.Context, path string) ([]byte, error) {
	document, err := Load(ctx, path)
	if err != nil {
		return nil, err
	}
	names := make(map[string]string)
	for name, schema := range document.Components.Schemas {
		if schema.RefPath() != nil {
			names[schema.RefPath().String()] = name
		}
	}
	document.InternalizeRefs(ctx, func(doc *openapi3.T, ref openapi3.ComponentRef) string {
		if name, ok := names[ref.RefPath().String()]; ok && ref.CollectionName() == "schemas" {
			return name
		}
		return openapi3.DefaultRefNameResolver(doc, ref)
	})
	return document.MarshalJSON()
}
