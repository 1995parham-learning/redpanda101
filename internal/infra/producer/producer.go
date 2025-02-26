package producer

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"

	"github.com/1995parham-teaching/redpanda101/internal/domain/model"
	"github.com/1995parham-teaching/redpanda101/internal/infra/constant"
	"github.com/twmb/franz-go/pkg/kgo"
)

type Producer struct {
	client *kgo.Client
}

func Provide(client *kgo.Client) *Producer {
	return &Producer{
		client: client,
	}
}

func (p *Producer) Produce(ctx context.Context, r model.Order) error {
	data, err := json.Marshal(r)
	if err != nil {
		return fmt.Errorf("converting order to json failed %w", err)
	}

	// nolint: mnd
	key := make([]byte, 8)
	binary.LittleEndian.PutUint64(key, r.ID)

	// nolint: exhaustruct
	record := &kgo.Record{
		Topic: constant.Topic,
		Value: data,
		Key:   key,
	}

	if err := p.client.ProduceSync(ctx, record).FirstErr(); err != nil {
		return fmt.Errorf("record had a produce error while synchronously producing: %w", err)
	}

	return nil
}
