package server

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-fuego/fuego"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func Provide(lc fx.Lifecycle, logger *zap.Logger) *fuego.Server {
	s := fuego.NewServer(
		fuego.WithAddr(":1378"),
	)

	lc.Append(
		fx.Hook{
			OnStart: func(_ context.Context) error {
				go func() {
					if err := s.Run(); !errors.Is(err, http.ErrServerClosed) {
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
