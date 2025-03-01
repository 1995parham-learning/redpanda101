package consumer

import (
	"github.com/1995parham-teaching/redpanda101/internal/infra/config"
	"github.com/1995parham-teaching/redpanda101/internal/infra/consumer"
	"github.com/1995parham-teaching/redpanda101/internal/infra/kafka"
	"github.com/1995parham-teaching/redpanda101/internal/infra/logger"
	"github.com/pterm/pterm"
	"github.com/pterm/pterm/putils"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

func main(_ consumer.Consumer) {
	area, _ := pterm.DefaultArea.WithCenter().Start()
	text, _ := pterm.DefaultBigText.WithLetters(putils.LettersFromString("Redpanda101")).Srender()
	area.Update(text)
}

func Register(root *cobra.Command) {
	root.AddCommand(
		//nolint: exhaustruct
		&cobra.Command{
			Use:   "consume",
			Short: "Consume orders from redpanda 🐼",
			Run: func(_ *cobra.Command, _ []string) {
				fx.New(
					fx.Provide(config.Provide),
					fx.Provide(logger.Provide),
					fx.WithLogger(func(logger *zap.Logger) fxevent.Logger {
						return &fxevent.ZapLogger{Logger: logger}
					}),
					fx.Provide(kafka.Provide),
					fx.Provide(consumer.Provide),
					fx.Invoke(main),
				).Run()
			},
		},
	)
}
