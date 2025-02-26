package config

import (
	"github.com/1995parham-teaching/redpanda101/internal/infra/kafka"
	"go.uber.org/fx"
)

// Default return default configuration.
func Default() Config {
	return Config{
		Kafka: kafka.Config{
			Seeds:         []string{"127.0.0.1:19092"},
			ConsumerGroup: "koochooloo-group",
		},
		Out: fx.Out{},
	}
}
