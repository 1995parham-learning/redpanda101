package producer

import (
	"context"
	"fmt"

	"github.com/1995parham-teaching/redpanda101/internal/domain/model"
	"github.com/1995parham-teaching/redpanda101/internal/infra/constant"
	"github.com/twmb/franz-go/pkg/kgo"
)

type Producer struct {
	client kgo.Client
}

func Provide() *Producer {
	return &Producer{}
}

func (p *Producer) Produce(ctx context.Context, r model.Order) error {
	record := &kgo.Record{
		Topic: constant.Topic,
	}

	if err := p.client.ProduceSync(ctx, record).FirstErr(); err != nil {
		return fmt.Errorf("record had a produce error while synchronously producing: %w", err)
	}

	return nil
}
