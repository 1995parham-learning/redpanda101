package kafka

import (
	"fmt"

	"github.com/twmb/franz-go/pkg/kgo"
)

func Provide(cfg Config) (*kgo.Client, error) {
	client, err := kgo.NewClient(
		kgo.SeedBrokers(cfg.Seeds...),
		kgo.ConsumerGroup(cfg.ConsumerGroup),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka client %w", err)
	}

	return client, nil
}
