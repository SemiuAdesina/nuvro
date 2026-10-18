package app

import (
	"context"
	"errors"

	"github.com/SemiuAdesina/nuvro/backend/internal/adapter/postgres"
	redisadapter "github.com/SemiuAdesina/nuvro/backend/internal/adapter/redis"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/config"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/telemetry"
	"github.com/SemiuAdesina/nuvro/backend/internal/usecase/system"
)

type dependencies struct {
	postgres  *postgres.Pool
	store     *postgres.Store
	redis     *redisadapter.Client
	telemetry *telemetry.Runtime
	probes    []system.Probe
}

func openDependencies(cfg config.Config) (*dependencies, error) {
	d := &dependencies{telemetry: telemetry.New()}
	var err error
	if cfg.Database.URL != "" {
		d.postgres, err = postgres.Open(context.Background(), cfg.Database.URL)
		if err != nil {
			_ = d.close(context.Background())
			return nil, err
		}
		d.store = postgres.NewStore(d.postgres.Pool)
		d.probes = append(d.probes, d.postgres)
	}
	if cfg.Redis.URL != "" {
		d.redis, err = redisadapter.Open(cfg.Redis.URL)
		if err != nil {
			_ = d.close(context.Background())
			return nil, err
		}
		d.probes = append(d.probes, d.redis)
	}
	return d, nil
}

func (d *dependencies) close(ctx context.Context) error {
	var err error
	if d.postgres != nil {
		d.postgres.Close()
	}
	if d.redis != nil {
		err = d.redis.Close()
	}
	return errors.Join(err, d.telemetry.Close(ctx))
}
