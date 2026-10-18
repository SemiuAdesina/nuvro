package plan

import (
	"context"

	"github.com/SemiuAdesina/nuvro/backend/internal/apperror"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/trip"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/cursor"
)

const cursorScope = "trips"
const maximumPageSize = 100
const defaultPageSize = 20

type TripLister interface {
	ListForMember(ctx context.Context, userID string, after *cursor.Position, limit int32) ([]trip.Trip, error)
}

type Page struct {
	Trips      []trip.Trip
	NextCursor string
	HasMore    bool
}

type Lister struct {
	repository TripLister
	cursors    cursor.Codec
}

func NewLister(repository TripLister, cursors cursor.Codec) *Lister {
	return &Lister{repository: repository, cursors: cursors}
}

func (l *Lister) List(ctx context.Context, principal user.Principal, token string, limit int) (Page, error) {
	if principal.UserID == "" {
		return Page{}, apperror.New(apperror.CodeUnauthenticated, "Authentication is required.")
	}
	size := pageSize(limit)
	var after *cursor.Position
	if token != "" {
		position, err := l.cursors.Decode(cursorScope, principal.UserID, token)
		if err != nil {
			return Page{}, apperror.New(apperror.CodeInvalidInput, "The cursor is not valid.")
		}
		after = &position
	}
	found, err := l.repository.ListForMember(ctx, principal.UserID, after, size+1)
	if err != nil {
		return Page{}, err
	}
	page := Page{Trips: found}
	if len(found) > int(size) {
		page.Trips = found[:size]
		page.HasMore = true
	}
	if !page.HasMore {
		return page, nil
	}
	last := page.Trips[len(page.Trips)-1]
	page.NextCursor, err = l.cursors.Encode(cursorScope, principal.UserID, cursor.Position{CreatedAt: last.CreatedAt, ID: last.ID})
	if err != nil {
		return Page{}, apperror.Wrap(apperror.CodeInternal, "", err)
	}
	return page, nil
}

func pageSize(limit int) int32 {
	if limit < 1 {
		return defaultPageSize
	}
	if limit > maximumPageSize {
		return maximumPageSize
	}
	return int32(limit)
}
