package consumer

import (
	"context"
	"encoding/json"
	"sync"
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
	wg     sync.WaitGroup
}

func Provide( // nolint: funlen
	lc fx.Lifecycle,
	client *kgo.Client,
	logger *zap.Logger,
	db *pgxpool.Pool,
	tracer *kotel.Tracer,
	tele telemetry.Telemetery,
) *Consumer {
	c := &Consumer{
		client: client,
		logger: logger,
		db:     db,
		tracer: tracer,
		metric: NewMetric(tele.MeterRegistry, tele.Namespace, tele.ServiceName),
		wg:     sync.WaitGroup{},
	}

	shutdown := make(chan struct{})

	client.AddConsumeTopics(constant.Topic)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			ctx = context.WithoutCancel(ctx)
			ctx, cancel := context.WithCancel(ctx)

			go func() {
				for {
					<-shutdown
					cancel()
				}
			}()

			go c.Consume(ctx)

			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("shutting down consumer gracefully")

			close(shutdown)

			done := make(chan struct{})

			go func() {
				c.wg.Wait()

				close(done)
			}()

			select {
			case <-done:
				logger.Info("consumer shutdown completed successfully")
			case <-ctx.Done():
				logger.Warn("consumer shutdown timed out, forcing shutdown")
			}

			return nil
		},
	})

	return c
}

func (c *Consumer) Consume(ctx context.Context) {
	ch := make(chan *kgo.Record, numberOfProcessors)

	for range numberOfProcessors {
		c.wg.Add(1)

		go c.process(ctx, ch)
	}

	c.logger.Info("consumer started", zap.Int("workers", numberOfProcessors))

	// Main consume loop
	for {
		select {
		case <-ctx.Done():
			c.logger.Info("consumer cancelled, stopping fetch loop")
			close(ch)

			return
		default:
		}

		fetches := c.client.PollFetches(ctx)

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

			select {
			case <-ctx.Done():
				c.logger.Info("consumer cancelled while sending record to workers")
				close(ch)

				return
			case ch <- record:
			}
		}
	}
}

func (c *Consumer) process(_ context.Context, ch <-chan *kgo.Record) {
	defer c.wg.Done()

	c.logger.Debug("worker started")

	for record := range ch {
		ctx, span := c.tracer.WithProcessSpan(record)

		c.metric.MessageDelay.Observe(time.Since(record.Timestamp).Seconds())

		var order model.Order

		err := json.Unmarshal(record.Value, &order)
		if err != nil {
			c.logger.Error("failed to parse an order from json", zap.Error(err), zap.ByteString("record", record.Value))
			span.RecordError(err)
			span.End()

			continue
		}

		c.logger.Info("new order received", zap.Any("order", order))

		start := time.Now()

		// nolint: contextcheck
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
			span.RecordError(err)
		}

		c.metric.DatabaseInsertionTime.Observe(time.Since(start).Seconds())

		span.End()
	}

	c.logger.Debug("worker stopped")
}
