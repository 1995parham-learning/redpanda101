package server

import (
	"context"
	"errors"
	"net/http"

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

	srv := &http.Server{
		Addr:    ":1378",
		Handler: e,
	}

	lc.Append(
		fx.Hook{
			OnStart: func(_ context.Context) error {
				go func() {
					err := srv.ListenAndServe()
					if !errors.Is(err, http.ErrServerClosed) {
						logger.Fatal("echo initiation failed", zap.Error(err))
					}
				}()

				return nil
			},
			OnStop: srv.Shutdown,
		},
	)

	return e
}
