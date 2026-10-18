package contract_test

import (
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/SemiuAdesina/nuvro/backend/internal/contract"
)

func released(t *testing.T) *openapi3.T {
	t.Helper()
	document, err := contract.Load(t.Context(), "../../../contracts/openapi/v1/openapi.yaml")
	if err != nil {
		t.Fatal(err)
	}
	return document
}

func TestEveryOperationIsNameableByAClientGenerator(t *testing.T) {
	seen := map[string]string{}
	for path, item := range released(t).Paths.Map() {
		for method, operation := range item.Operations() {
			if operation.OperationID == "" {
				t.Fatalf("%s %s has no operationId, so a generated client cannot name it", method, path)
			}
			if previous, clash := seen[operation.OperationID]; clash {
				t.Fatalf("operationId %q is used by both %s and %s %s",
					operation.OperationID, previous, method, path)
			}
			seen[operation.OperationID] = method + " " + path
		}
	}
}

func TestEveryOperationDeclaresAFailureShape(t *testing.T) {
	for path, item := range released(t).Paths.Map() {
		for method, operation := range item.Operations() {
			if operation.Responses == nil || operation.Responses.Default() == nil {
				t.Fatalf("%s %s declares no default problem response, so failures are untyped for clients",
					method, path)
			}
		}
	}
}

func TestEveryResponseBodyIsJSON(t *testing.T) {
	for path, item := range released(t).Paths.Map() {
		for method, operation := range item.Operations() {
			assertJSONOnly(t, method, path, operation)
		}
	}
}

func assertJSONOnly(t *testing.T, method, path string, operation *openapi3.Operation) {
	t.Helper()
	if body := operation.RequestBody; body != nil && body.Value != nil {
		assertMediaTypes(t, method, path, "request", body.Value.Content)
	}
	for status, response := range operation.Responses.Map() {
		if response.Value == nil {
			continue
		}
		assertMediaTypes(t, method, path, status, response.Value.Content)
	}
}

func assertMediaTypes(t *testing.T, method, path, label string, content openapi3.Content) {
	t.Helper()
	for mediaType := range content {
		if mediaType == "application/json" || mediaType == "application/problem+json" {
			continue
		}
		t.Fatalf("%s %s %s offers %q, which a generated client would have to special case",
			method, path, label, mediaType)
	}
}

func TestPathParametersAreDeclaredAndRequired(t *testing.T) {
	for path, item := range released(t).Paths.Map() {
		for _, name := range placeholders(path) {
			assertPathParameter(t, path, item, name)
		}
	}
}

func assertPathParameter(t *testing.T, path string, item *openapi3.PathItem, name string) {
	t.Helper()
	for method, operation := range item.Operations() {
		found := parameterFor(item.Parameters, name) || parameterFor(operation.Parameters, name)
		if !found {
			t.Fatalf("%s %s takes {%s} in its path but never declares it", method, path, name)
		}
	}
}

func parameterFor(parameters openapi3.Parameters, name string) bool {
	for _, parameter := range parameters {
		if parameter.Value == nil || parameter.Value.Name != name || parameter.Value.In != "path" {
			continue
		}
		if parameter.Value.Required {
			return true
		}
	}
	return false
}

func placeholders(path string) []string {
	var names []string
	for _, segment := range strings.Split(path, "/") {
		if strings.HasPrefix(segment, "{") && strings.HasSuffix(segment, "}") {
			names = append(names, strings.Trim(segment, "{}"))
		}
	}
	return names
}
