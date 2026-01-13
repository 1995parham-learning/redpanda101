package producer

import (
	"context"
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
	metric *Metric
}

func Provide(client *kgo.Client, tel telemetry.Telemetry) *Producer {
	return &Producer{
		client: client,
		tracer: tel.TraceProvider.Tracer("producer"),
		metric: NewMetric(tel.MeterRegistry, tel.Namespace, tel.ServiceName),
	}
}

func (p *Producer) Produce(ctx context.Context, r model.Order) error {
	ctx, span := p.tracer.Start(ctx, "produce.order", trace.WithSpanKind(trace.SpanKindProducer))
	defer span.End()

	start := time.Now()

	data, err := json.Marshal(r)
	if err != nil {
		span.RecordError(err)
		p.metric.ProduceErrors.Inc()

		return fmt.Errorf("converting order to json failed %w", err)
	}

	// nolint: exhaustruct
	record := &kgo.Record{
		Topic:     constant.Topic,
		Value:     data,
		Key:       []byte(r.ID),
		Timestamp: time.Now(),
	}

	err = p.client.ProduceSync(ctx, record).FirstErr()
	if err != nil {
		p.metric.ProduceErrors.Inc()

		return fmt.Errorf("record had a produce error while synchronously producing: %w", err)
	}

	p.metric.ProduceLatency.Observe(time.Since(start).Seconds())
	p.metric.MessagesProduced.Inc()

	return nil
}
