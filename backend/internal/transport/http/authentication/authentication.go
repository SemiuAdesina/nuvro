package authentication

import (
	"context"
	"errors"
	"strings"

	"github.com/getkin/kin-openapi/openapi3filter"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
)

type Authenticator interface {
	Authenticate(context.Context, string) (user.Principal, error)
}

type principalKey struct{}

func From(ctx context.Context) (user.Principal, bool) {
	principal, ok := ctx.Value(principalKey{}).(user.Principal)
	return principal, ok && principal.UserID != ""
}

func Validator(authenticator Authenticator) openapi3filter.AuthenticationFunc {
	return func(ctx context.Context, input *openapi3filter.AuthenticationInput) error {
		if authenticator == nil {
			return errors.New("authentication required")
		}
		request := input.RequestValidationInput.Request
		fields := strings.Fields(request.Header.Get("Authorization"))
		if len(fields) != 2 || !strings.EqualFold(fields[0], "Bearer") || len(fields[1]) > 8192 {
			return errors.New("authentication required")
		}
		principal, err := authenticator.Authenticate(ctx, fields[1])
		if err != nil || principal.UserID == "" || principal.Provider == "" || principal.Subject == "" {
			return errors.New("authentication required")
		}
		*request = *request.WithContext(context.WithValue(ctx, principalKey{}, principal))
		return nil
	}
}
