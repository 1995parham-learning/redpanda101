package migrate

import (
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

	if err := m.Up(); err != nil {
		logger.Fatal("applying migration failed", zap.Error(err))
	}

	_ = sh.Shutdown()
}

func Register(root *cobra.Command) {
	root.AddCommand(
		//nolint: exhaustruct
		&cobra.Command{
			Use:   "migrate",
			Short: "Applying migration",
			Run: func(_ *cobra.Command, _ []string) {
				fx.New(
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
