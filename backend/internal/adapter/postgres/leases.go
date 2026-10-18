package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Lease struct {
	pool *pgxpool.Pool
	name string
	conn *pgxpool.Conn
}

func NewLease(pool *pgxpool.Pool, name string) *Lease {
	return &Lease{pool: pool, name: name}
}

func (l *Lease) Acquire(ctx context.Context) (bool, error) {
	if l == nil || l.pool == nil {
		return true, nil
	}
	conn, err := l.pool.Acquire(ctx)
	if err != nil {
		return false, databaseError(err)
	}
	var held bool
	row := conn.QueryRow(ctx, "SELECT pg_try_advisory_lock(hashtext($1)::bigint)", l.name)
	if err := row.Scan(&held); err != nil {
		conn.Release()
		return false, databaseError(err)
	}
	if !held {
		conn.Release()
		return false, nil
	}
	l.conn = conn
	return true, nil
}

func (l *Lease) Release(ctx context.Context) {
	if l == nil || l.conn == nil {
		return
	}
	_, _ = l.conn.Exec(ctx, "SELECT pg_advisory_unlock(hashtext($1)::bigint)", l.name)
	l.conn.Release()
	l.conn = nil
}
