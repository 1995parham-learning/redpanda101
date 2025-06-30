package database

import (
	"context"
	"fmt"

	"github.com/1995parham-teaching/redpanda101/internal/infra/constant"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
)

func Provide(lc fx.Lifecycle, cfg Config) (*pgxpool.Pool, error) {
	ctx := context.Background()

	ctx, done := context.WithTimeout(ctx, constant.PingTimeout)
	defer done()

	pool, err := pgxpool.New(ctx, cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("database connection failed %w", err)
	}

	ctx = context.Background()

	ctx, done = context.WithTimeout(ctx, constant.PingTimeout)
	defer done()

	err = pool.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf("database ping failed %w", err)
	}

	lc.Append(fx.StopHook(func() {
		pool.Close()
	}))

	return pool, nil
}
