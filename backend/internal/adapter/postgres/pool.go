package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Pool struct{ *pgxpool.Pool }

func Open(ctx context.Context, url string) (*Pool, error) {
	configuration, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, errors.New("postgres: invalid configuration")
	}
	configuration.ConnConfig.ConnectTimeout = 2 * time.Second
	configuration.ConnConfig.RuntimeParams["timezone"] = "UTC"
	configuration.MaxConns = 10
	pool, err := pgxpool.NewWithConfig(ctx, configuration)
	if err != nil {
		return nil, errors.New("postgres: initialization failed")
	}
	return &Pool{Pool: pool}, nil
}

func (*Pool) Name() string                      { return "postgres" }
func (p *Pool) Check(ctx context.Context) error { return p.Ping(ctx) }
