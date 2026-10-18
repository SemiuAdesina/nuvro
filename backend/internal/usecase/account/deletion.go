package account

import (
	"context"
	"reflect"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/apperror"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/retention"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
)

type Deletions interface {
	RequestDeletion(context.Context, retention.Deletion) (retention.Deletion, error)
	Deletion(context.Context, string) (retention.Deletion, error)
}

type Identifiers interface{ Generate() (string, error) }

type Revoker interface {
	RevokeFor(context.Context, string) (int, error)
}

type SessionRevoker interface {
	RevokeAll(context.Context, string) error
}

type Eraser struct {
	deletions   Deletions
	identifiers Identifiers
	revoker     Revoker
	sessions    SessionRevoker
	timer       clock.Clock
}

func (e *Eraser) WithSessions(sessions SessionRevoker) *Eraser {
	if sessions == nil || reflect.ValueOf(sessions).IsNil() {
		return e
	}
	e.sessions = sessions
	return e
}

func NewEraser(deletions Deletions, identifiers Identifiers, timer clock.Clock) *Eraser {
	return &Eraser{deletions: deletions, identifiers: identifiers, timer: timer}
}

func (e *Eraser) WithLinker(revoker Revoker) *Eraser {
	if revoker == nil || reflect.ValueOf(revoker).IsNil() {
		return e
	}
	e.revoker = revoker
	return e
}

func (e *Eraser) Request(ctx context.Context, principal user.Principal, reason string) (retention.Deletion, error) {
	if principal.UserID == "" {
		return retention.Deletion{}, apperror.New(apperror.CodeUnauthenticated, "Authentication is required.")
	}
	deletionID, err := e.identifiers.Generate()
	if err != nil {
		return retention.Deletion{}, apperror.New(apperror.CodeInternal, "The deletion could not be started.")
	}
	requested, err := retention.Request(deletionID, principal.UserID, reason, e.timer.Now())
	if err != nil {
		return retention.Deletion{}, err
	}
	accepted, err := e.deletions.RequestDeletion(ctx, requested)
	if err != nil {
		return retention.Deletion{}, err
	}
	if e.sessions != nil {
		_ = e.sessions.RevokeAll(ctx, principal.UserID)
	}
	e.revoke(ctx, principal.UserID)
	return accepted, nil
}

func (e *Eraser) revoke(ctx context.Context, userID string) {
	if e.revoker == nil {
		return
	}
	_, _ = e.revoker.RevokeFor(ctx, userID)
}

func (e *Eraser) Status(ctx context.Context, principal user.Principal) (retention.Deletion, error) {
	if principal.UserID == "" {
		return retention.Deletion{}, apperror.New(apperror.CodeUnauthenticated, "Authentication is required.")
	}
	return e.deletions.Deletion(ctx, principal.UserID)
}

type Purger interface {
	DueDeletions(context.Context, time.Time, int) ([]retention.Deletion, error)
	Purge(context.Context, retention.Deletion, time.Time) (retention.Deletion, error)
}

type Sweeper interface {
	Sweep(context.Context, retention.Rule, time.Time) (int64, error)
}
