package server

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/1995parham-teaching/redpanda101/internal/infra/http/controller"
	"github.com/1995parham-teaching/redpanda101/internal/infra/producer"
	"github.com/labstack/echo/v5"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func Provide(lc fx.Lifecycle, logger *zap.Logger, p *producer.Producer) *echo.Echo {
	e := echo.New()

	controller.Order{
		Producer: p,
	}.Register(e)

	// Use echo's StartConfig (see https://echo.labstack.com/docs/start-server)
	// instead of constructing a bare http.Server. BeforeServeFunc lets us set
	// ReadHeaderTimeout to mitigate Slowloris (gosec G112), and graceful
	// shutdown is driven by cancelling the context passed to sc.Start.
	const readHeaderTimeout = 5 * time.Second

	sc := echo.StartConfig{ //nolint:exhaustruct
		Address: ":1378",
		BeforeServeFunc: func(s *http.Server) error {
			s.ReadHeaderTimeout = readHeaderTimeout

			return nil
		},
	}

	ctx, cancel := context.WithCancel(context.Background())

	lc.Append(
		fx.Hook{
			OnStart: func(_ context.Context) error {
				go func() {
					if err := sc.Start(ctx, e); err != nil && !errors.Is(err, http.ErrServerClosed) {
						logger.Fatal("echo initiation failed", zap.Error(err))
					}
				}()

				return nil
			},
			OnStop: func(_ context.Context) error {
				cancel()

				return nil
			},
		},
	)

	return e
}
