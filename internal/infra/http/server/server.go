package server

import (
	"context"
	"errors"
	"net/http"

	"github.com/1995parham-teaching/redpanda101/internal/infra/http/controller"
	"github.com/1995parham-teaching/redpanda101/internal/infra/producer"
	"github.com/go-fuego/fuego"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func Provide(lc fx.Lifecycle, logger *zap.Logger, p *producer.Producer) *fuego.Server {
	s := fuego.NewServer(
		fuego.WithAddr(":1378"),
	)

	controller.Order{
		Producer: p,
	}.Register(s)

	lc.Append(
		fx.Hook{
			OnStart: func(_ context.Context) error {
				go func() {
					err := s.Run()
					if !errors.Is(err, http.ErrServerClosed) {
						logger.Fatal("echo initiation failed", zap.Error(err))
					}
				}()

				return nil
			},
			OnStop: s.Shutdown,
		},
	)

	return s
}
