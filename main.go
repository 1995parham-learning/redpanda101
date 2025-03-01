package main

import (
	"github.com/1995parham-teaching/redpanda101/internal/infra/config"
	"github.com/1995parham-teaching/redpanda101/internal/infra/http/server"
	"github.com/1995parham-teaching/redpanda101/internal/infra/kafka"
	"github.com/1995parham-teaching/redpanda101/internal/infra/logger"
	"github.com/1995parham-teaching/redpanda101/internal/infra/producer"
	"github.com/go-fuego/fuego"
	"github.com/pterm/pterm"
	"github.com/pterm/pterm/putils"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

func main() {
	fx.New(
		fx.Provide(config.Provide),
		fx.Provide(logger.Provide),
		fx.WithLogger(func(logger *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: logger}
		}),
		fx.Provide(kafka.Provide),
		fx.Provide(producer.Provide),
		fx.Provide(server.Provide),
		fx.Invoke(func(_ *fuego.Server) {
			area, _ := pterm.DefaultArea.WithCenter().Start()
			text, _ := pterm.DefaultBigText.WithLetters(putils.LettersFromString("Redpanda101")).Srender()
			area.Update(text)
		}),
	).Run()
}
