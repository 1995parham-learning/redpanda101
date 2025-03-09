package config

import (
	"github.com/1995parham-teaching/redpanda101/internal/infra/database"
	"github.com/1995parham-teaching/redpanda101/internal/infra/kafka"
	"github.com/1995parham-teaching/redpanda101/internal/infra/logger"
	"github.com/1995parham-teaching/redpanda101/internal/infra/telemetry"
	"go.uber.org/fx"
)

// Default return default configuration.
func Default() Config {
	return Config{
		Logger: logger.Config{
			Level: "info",
		},
		Database: database.Config{
			URL: "postgres://username:password@127.0.0.1:5432/redpanda",
		},
		Kafka: kafka.Config{
			Seeds:         []string{"127.0.0.1:19092"},
			ConsumerGroup: "koochooloo-group",
		},
		Telemetry: telemetry.Config{
			Trace: telemetry.Trace{
				Enabled:  true,
				Endpoint: "127.0.0.1:4317",
			},
			Meter: telemetry.Meter{
				Address: ":8080",
				Enabled: true,
			},
			Namespace:   "1995parham-learning",
			ServiceName: "redpanda101",
		},
		Out: fx.Out{},
	}
}
