package migrate

import (
	"errors"

	"github.com/1995parham-teaching/redpanda101/internal/infra/config"
	"github.com/1995parham-teaching/redpanda101/internal/infra/database"
	"github.com/1995parham-teaching/redpanda101/internal/infra/logger"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file" // required by go-migrate
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

func main(sh fx.Shutdowner, logger *zap.Logger, db *pgxpool.Pool) {
	conn := stdlib.OpenDBFromPool(db)

	// nolint: exhaustruct
	driver, err := postgres.WithInstance(conn, &postgres.Config{})
	if err != nil {
		logger.Fatal("creating database instance failed", zap.Error(err))
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://./migrations",
		"postgres", driver,
	)
	if err != nil {
		logger.Fatal("loading migration failed", zap.Error(err))
	}

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		logger.Fatal("applying migration failed", zap.Error(err))
	}

	if errors.Is(err, migrate.ErrNoChange) {
		logger.Info("no new migrations to apply")
	} else {
		logger.Info("migrations applied successfully")
	}

	_ = sh.Shutdown()
}

func Register(root *cobra.Command) {
	root.AddCommand(
		//nolint: exhaustruct
		&cobra.Command{
			Use:   "migrate",
			Short: "Applying migration",
			Run: func(cmd *cobra.Command, _ []string) {
				path := cmd.Flag("config").Value.String()

				fx.New(
					fx.Supply(config.Path(path)),
					fx.Provide(config.Provide),
					fx.Provide(logger.Provide),
					fx.WithLogger(func(logger *zap.Logger) fxevent.Logger {
						return &fxevent.ZapLogger{Logger: logger}
					}),
					fx.Provide(database.Provide),
					fx.Invoke(main),
				).Run()
			},
		},
	)
}
