package account

import (
	"context"

	"github.com/SemiuAdesina/nuvro/backend/internal/apperror"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/id"
)

type Directory interface {
	FindUserByIdentity(ctx context.Context, provider, subject string) (string, string, error)
	CreateAccount(context.Context, user.Account) error
}

type Transaction interface {
	Within(context.Context, func(context.Context) error) error
}

type Resolver struct {
	directory   Directory
	transaction Transaction
	identifiers id.Generator
}

func NewResolver(directory Directory, transaction Transaction, identifiers id.Generator) *Resolver {
	return &Resolver{directory: directory, transaction: transaction, identifiers: identifiers}
}

func (r *Resolver) Resolve(ctx context.Context, identity user.Identity) (user.Principal, error) {
	if identity.Provider == "" || identity.Subject == "" {
		return user.Principal{}, apperror.New(apperror.CodeUnauthenticated, "Authentication is required.")
	}
	principal, err := r.existing(ctx, identity)
	if err != nil || principal.UserID != "" {
		return principal, err
	}
	if enrolErr := r.enrol(ctx, identity); enrolErr != nil && apperror.CodeOf(enrolErr) != apperror.CodeConflict {
		return user.Principal{}, enrolErr
	}
	principal, err = r.existing(ctx, identity)
	if err != nil {
		return user.Principal{}, err
	}
	if principal.UserID == "" {
		return user.Principal{}, apperror.New(apperror.CodeInternal, "")
	}
	return principal, nil
}

func (r *Resolver) existing(ctx context.Context, identity user.Identity) (user.Principal, error) {
	userID, status, err := r.directory.FindUserByIdentity(ctx, identity.Provider, identity.Subject)
	if err != nil || userID == "" {
		return user.Principal{}, err
	}
	if !user.IsActive(status) {
		return user.Principal{}, apperror.New(apperror.CodeForbidden, "This account cannot be used.")
	}
	return user.Principal{UserID: userID, Subject: identity.Subject, Provider: identity.Provider}, nil
}

func (r *Resolver) enrol(ctx context.Context, identity user.Identity) error {
	userID, err := r.identifiers.Generate()
	if err != nil {
		return apperror.Wrap(apperror.CodeInternal, "", err)
	}
	identityID, err := r.identifiers.Generate()
	if err != nil {
		return apperror.Wrap(apperror.CodeInternal, "", err)
	}
	account, err := user.NewAccount(userID, identityID, identity)
	if err != nil {
		return apperror.Wrap(apperror.CodeInternal, "", err)
	}
	return r.transaction.Within(ctx, func(txctx context.Context) error {
		return r.directory.CreateAccount(txctx, account)
	})
}
