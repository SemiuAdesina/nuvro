package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	queries "github.com/SemiuAdesina/nuvro/backend/internal/adapter/postgres/sqlc"
	"github.com/SemiuAdesina/nuvro/backend/internal/apperror"
)

const uniqueViolation = "23505"

type transactionKey struct{}
type transactionContext struct {
	owner *Store
	tx    pgx.Tx
}
type Store struct{ pool *pgxpool.Pool }

func NewStore(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

func (s *Store) Within(ctx context.Context, operation func(context.Context) error) error {
	if ctx.Value(transactionKey{}) != nil {
		return errors.New("postgres: nested transaction rejected")
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return databaseError(err)
	}
	defer func() {
		cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), time.Second)
		defer cancel()
		_ = tx.Rollback(cleanup)
	}()
	txctx := context.WithValue(ctx, transactionKey{}, transactionContext{owner: s, tx: tx})
	if err := operation(txctx); err != nil {
		return err
	}
	return databaseError(tx.Commit(ctx))
}

func (s *Store) queries(ctx context.Context) *queries.Queries {
	transaction, ok := ctx.Value(transactionKey{}).(transactionContext)
	if ok && transaction.owner == s {
		return queries.New(transaction.tx)
	}
	return queries.New(s.pool)
}

func databaseError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return apperror.New(apperror.CodeNotFound, "Resource not found.")
	}
	var database *pgconn.PgError
	if errors.As(err, &database) && database.Code == uniqueViolation {
		return apperror.Wrap(apperror.CodeConflict, "The resource already exists.", err)
	}
	return apperror.Wrap(apperror.CodeInternal, "", err)
}
