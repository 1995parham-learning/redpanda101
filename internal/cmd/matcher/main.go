package matcher

import (
	"github.com/1995parham-teaching/redpanda101/internal/infra/config"
	"github.com/1995parham-teaching/redpanda101/internal/infra/database"
	"github.com/1995parham-teaching/redpanda101/internal/infra/logger"
	"github.com/1995parham-teaching/redpanda101/internal/infra/matcher"
	"github.com/1995parham-teaching/redpanda101/internal/infra/telemetry"
	"github.com/pterm/pterm"
	"github.com/pterm/pterm/putils"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

func main(_ *matcher.Matcher) {
	area, _ := pterm.DefaultArea.WithCenter().Start()
	text, _ := pterm.DefaultBigText.WithLetters(putils.LettersFromString("Redpanda101")).Srender()
	area.Update(text)
}

func Register(root *cobra.Command) {
	root.AddCommand(
		//nolint: exhaustruct
		&cobra.Command{
			Use:   "match",
			Short: "Match orders into trades using an in-memory order book 🐼",
			Run: func(cmd *cobra.Command, _ []string) {
				path := cmd.Flag("config").Value.String()

				fx.New(
					fx.Supply(config.Path(path)),
					fx.Provide(config.Provide),
					fx.Provide(logger.Provide),
					fx.WithLogger(func(logger *zap.Logger) fxevent.Logger {
						return &fxevent.ZapLogger{Logger: logger}
					}),
					fx.Provide(telemetry.Provide),
					fx.Provide(database.Provide),
					fx.Provide(matcher.Provide),
					fx.Invoke(main),
				).Run()
			},
		},
	)
}
