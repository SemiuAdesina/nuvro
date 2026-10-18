package middleware

import (
	"context"
	"errors"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	validator "github.com/oapi-codegen/nethttp-middleware"

	"github.com/SemiuAdesina/nuvro/backend/internal/transport/http/httperror"
)

func Validate(document *openapi3.T, authenticate openapi3filter.AuthenticationFunc) func(http.Handler) http.Handler {
	return validator.OapiRequestValidatorWithOptions(document, &validator.Options{
		Options:              openapi3filter.Options{AuthenticationFunc: authenticate},
		ErrorHandlerWithOpts: validationError,
	})
}

func validationError(_ context.Context, err error, w http.ResponseWriter, _ *http.Request, options validator.ErrorHandlerOpts) {
	status := options.StatusCode
	code := "validation_error"
	var security *openapi3filter.SecurityRequirementsError
	var tooLarge *http.MaxBytesError
	switch {
	case errors.As(err, &security):
		status, code = 401, "authentication_required"
	case errors.As(err, &tooLarge):
		status, code = 413, "payload_too_large"
	case status == 404:
		code = "not_found"
	case status == 405:
		code = "method_not_allowed"
	case status == 415:
		code = "unsupported_media_type"
	}
	httperror.WriteStatus(w, status, code, http.StatusText(status))
}
