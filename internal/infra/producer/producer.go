package producer

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"time"

	"github.com/1995parham-teaching/redpanda101/internal/domain/model"
	"github.com/1995parham-teaching/redpanda101/internal/infra/constant"
	"github.com/1995parham-teaching/redpanda101/internal/infra/telemetry"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.opentelemetry.io/otel/trace"
)

type Producer struct {
	client *kgo.Client
	tracer trace.Tracer
}

func Provide(client *kgo.Client, tel telemetry.Telemetery) *Producer {
	return &Producer{
		client: client,
		tracer: tel.TraceProvider.Tracer("producer"),
	}
}

func (p *Producer) Produce(ctx context.Context, r model.Order) error {
	ctx, span := p.tracer.Start(ctx, "produce.order", trace.WithSpanKind(trace.SpanKindProducer))
	defer span.End()

	data, err := json.Marshal(r)
	if err != nil {
		span.RecordError(err)

		return fmt.Errorf("converting order to json failed %w", err)
	}

	// nolint: mnd
	key := make([]byte, 8)
	binary.LittleEndian.PutUint64(key, r.ID)

	// nolint: exhaustruct
	record := &kgo.Record{
		Topic:     constant.Topic,
		Value:     data,
		Key:       key,
		Headers:   []kgo.RecordHeader{},
		Timestamp: time.Now(),
	}

	if err := p.client.ProduceSync(ctx, record).FirstErr(); err != nil {
		return fmt.Errorf("record had a produce error while synchronously producing: %w", err)
	}

	return nil
}
