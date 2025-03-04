package consumer

import (
	"context"
	"encoding/json"

	"github.com/1995parham-teaching/redpanda101/internal/domain/model"
	"github.com/1995parham-teaching/redpanda101/internal/infra/constant"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Consumer struct {
	client *kgo.Client
	logger *zap.Logger
	db     *pgxpool.Pool
}

func Provide(lc fx.Lifecycle, client *kgo.Client, logger *zap.Logger, db *pgxpool.Pool) Consumer {
	c := Consumer{
		client: client,
		logger: logger,
		db:     db,
	}

	client.AddConsumeTopics(constant.Topic)

	lc.Append(fx.StartHook(func() {
		go c.Consume()
	}))

	return c
}

func (c Consumer) Consume() {
	ctx := context.Background()

	for {
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

			var order model.Order
			if err := json.Unmarshal(record.Value, &order); err != nil {
				c.logger.Error("failed to parse an order from json", zap.Error(err), zap.ByteString("record", record.Value))
			}

			c.logger.Info("new order received", zap.Any("order", order))

			if _, err := c.db.Exec(
				context.Background(),
				"INSERT INTO orders (description, src_currency, dst_currency, channel) VALUES ($1, $2, $3, $4)",
				order.Description,
				order.SrcCurrency,
				order.DstCurrency,
				order.Channel,
			); err != nil {
				c.logger.Error("database insertion failed", zap.Error(err))
			}
		}
	}
}
