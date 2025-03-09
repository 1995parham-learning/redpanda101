package producer

import (
	"github.com/1995parham-teaching/redpanda101/internal/infra/config"
	"github.com/1995parham-teaching/redpanda101/internal/infra/http/server"
	"github.com/1995parham-teaching/redpanda101/internal/infra/kafka"
	"github.com/1995parham-teaching/redpanda101/internal/infra/logger"
	"github.com/1995parham-teaching/redpanda101/internal/infra/producer"
	"github.com/1995parham-teaching/redpanda101/internal/infra/telemetry"
	"github.com/go-fuego/fuego"
	"github.com/pterm/pterm"
	"github.com/pterm/pterm/putils"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

func main(_ *fuego.Server) {
	area, _ := pterm.DefaultArea.WithCenter().Start()
	text, _ := pterm.DefaultBigText.WithLetters(putils.LettersFromString("Redpanda101")).Srender()
	area.Update(text)
}

func Register(root *cobra.Command) {
	root.AddCommand(
		//nolint: exhaustruct
		&cobra.Command{
			Use:   "produce",
			Short: "Create orders from web and put them in redpanda 🐼",
			Run: func(_ *cobra.Command, _ []string) {
				fx.New(
					fx.Provide(config.Provide),
					fx.Provide(logger.Provide),
					fx.WithLogger(func(logger *zap.Logger) fxevent.Logger {
						return &fxevent.ZapLogger{Logger: logger}
					}),
					fx.Provide(telemetry.Provide),
					fx.Provide(kafka.Provide),
					fx.Provide(producer.Provide),
					fx.Provide(server.Provide),
					fx.Invoke(main),
				).Run()
			},
		},
	)
}
