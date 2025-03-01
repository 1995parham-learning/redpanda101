package kafka

import (
	"context"
	"fmt"

	"github.com/1995parham-teaching/redpanda101/internal/infra/constant"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/fx"
)

func Provide(lc fx.Lifecycle, cfg Config) (*kgo.Client, error) {
	client, err := kgo.NewClient(
		kgo.SeedBrokers(cfg.Seeds...),
		kgo.ConsumerGroup(cfg.ConsumerGroup),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka client %w", err)
	}

	ctx := context.Background()

	ctx, done := context.WithTimeout(ctx, constant.PingTimeout)
	defer done()

	if err := client.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping kafka cluster %w", err)
	}

	lc.Append(fx.StopHook(func() {
		client.Close()
	}))

	return client, nil
}
