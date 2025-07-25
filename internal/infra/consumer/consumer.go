package consumer

import (
	"context"
	"encoding/json"
	"time"

	"github.com/1995parham-teaching/redpanda101/internal/domain/model"
	"github.com/1995parham-teaching/redpanda101/internal/infra/constant"
	"github.com/1995parham-teaching/redpanda101/internal/infra/telemetry"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/plugin/kotel"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const numberOfProcessors = 20

type Consumer struct {
	client *kgo.Client
	logger *zap.Logger
	db     *pgxpool.Pool
	tracer *kotel.Tracer
	metric *Metric
}

func Provide(
	lc fx.Lifecycle,
	client *kgo.Client,
	logger *zap.Logger,
	db *pgxpool.Pool,
	tracer *kotel.Tracer,
	tele telemetry.Telemetery,
) Consumer {
	c := Consumer{
		client: client,
		logger: logger,
		db:     db,
		tracer: tracer,
		metric: NewMetric(tele.MeterRegistry, tele.Namespace, tele.ServiceName),
	}

	client.AddConsumeTopics(constant.Topic)

	lc.Append(fx.StartHook(func() {
		go c.Consume()
	}))

	return c
}

func (c Consumer) Consume() {
	ch := make(chan *kgo.Record)

	for range numberOfProcessors {
		go c.process(ch)
	}

	for {
		fetches := c.client.PollFetches(context.Background())

		if errs := fetches.Errors(); len(errs) > 0 {
			for _, err := range errs {
				c.logger.Error(
					"failed to fetch messages from kafka",
					zap.Error(err.Err),
					zap.String("topic", err.Topic),
					zap.Int32("partition", err.Partition),
				)
			}
		}

		iter := fetches.RecordIter()
		for !iter.Done() {
			record := iter.Next()
			ch <- record
		}
	}
}

func (c Consumer) process(ch <-chan *kgo.Record) {
	for record := range ch {
		ctx, span := c.tracer.WithProcessSpan(record)

		c.metric.MessageDelay.Observe(time.Since(record.Timestamp).Seconds())

		var order model.Order

		err := json.Unmarshal(record.Value, &order)
		if err != nil {
			c.logger.Error("failed to parse an order from json", zap.Error(err), zap.ByteString("record", record.Value))
		}

		c.logger.Info("new order received", zap.Any("order", order))

		start := time.Now()

		_, err = c.db.Exec(
			ctx,
			"INSERT INTO orders (description, src_currency, dst_currency, channel) VALUES ($1, $2, $3, $4)",
			order.Description,
			order.SrcCurrency,
			order.DstCurrency,
			order.Channel,
		)
		if err != nil {
			c.logger.Error("database insertion failed", zap.Error(err))
		}

		c.metric.DatabaseInsertionTime.Observe(time.Since(start).Seconds())

		span.End()
	}
}
