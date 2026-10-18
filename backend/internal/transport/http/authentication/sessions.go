package authentication

import (
	"context"
	"reflect"

	"github.com/SemiuAdesina/nuvro/backend/internal/apperror"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
)

type Verifier interface {
	Verify(context.Context, string) (user.Identity, error)
}

type Accounts interface {
	Resolve(context.Context, user.Identity) (user.Principal, error)
}

type Tokens interface {
	Authenticate(context.Context, string) (user.Principal, bool, error)
}

type Sessions struct {
	verifier Verifier
	accounts Accounts
	tokens   Tokens
}

func NewSessions(verifier Verifier, accounts Accounts) *Sessions {
	if verifier == nil || accounts == nil {
		return nil
	}
	return &Sessions{verifier: verifier, accounts: accounts}
}

func (s *Sessions) WithTokens(tokens Tokens) *Sessions {
	if tokens == nil || reflect.ValueOf(tokens).IsNil() {
		return s
	}
	s.tokens = tokens
	return s
}

func (s *Sessions) Authenticate(ctx context.Context, credential string) (user.Principal, error) {
	if s.tokens != nil {
		principal, handled, err := s.tokens.Authenticate(ctx, credential)
		if handled {
			return principal, err
		}
	}
	identity, err := s.verifier.Verify(ctx, credential)
	if err != nil {
		return user.Principal{}, apperror.New(apperror.CodeUnauthenticated, "Authentication is required.")
	}
	principal, err := s.accounts.Resolve(ctx, identity)
	if err != nil {
		return user.Principal{}, err
	}
	return principal, nil
}
