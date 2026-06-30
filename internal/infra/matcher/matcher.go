package matcher

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/1995parham-teaching/redpanda101/internal/domain/model"
	"github.com/1995parham-teaching/redpanda101/internal/domain/orderbook"
	"github.com/1995parham-teaching/redpanda101/internal/infra/constant"
	"github.com/1995parham-teaching/redpanda101/internal/infra/kafka"
	"github.com/1995parham-teaching/redpanda101/internal/infra/telemetry"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/plugin/kotel"
	"go.opentelemetry.io/otel/propagation"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// snapshotDepth is how many price levels per side we log after each order.
const snapshotDepth = 5

// Matcher is the single-writer matching service. It consumes the orders log
// sequentially, feeds each order into an in-memory order book (materialised
// from Redpanda — the orders topic is the source of truth), and emits the
// resulting trades to the trades topic while persisting them to PostgreSQL.
//
// Ordering matters: orders are processed one at a time, never on a worker pool,
// so the book sees them in the same order Redpanda stored them. The producer
// keys orders by symbol, so each market is totally ordered on its partition.
type Matcher struct {
	client *kgo.Client
	engine *orderbook.Engine
	logger *zap.Logger
	db     *pgxpool.Pool
	metric *Metric
	done   chan struct{}
}

func Provide( //nolint:funlen
	lc fx.Lifecycle,
	cfg kafka.Config,
	logger *zap.Logger,
	db *pgxpool.Pool,
	tele telemetry.Telemetry,
) (*Matcher, error) {
	client, err := newClient(cfg, tele)
	if err != nil {
		return nil, err
	}

	m := &Matcher{
		client: client,
		engine: orderbook.NewEngine(),
		logger: logger,
		db:     db,
		metric: NewMetric(tele.MeterRegistry, tele.Namespace, tele.ServiceName),
		done:   make(chan struct{}),
	}

	shutdown := make(chan struct{})

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			ctx = context.WithoutCancel(ctx)
			ctx, cancel := context.WithCancel(ctx) //nolint:gosec // cancelled by the shutdown goroutine

			go func() {
				<-shutdown
				cancel()
			}()

			go m.run(ctx)

			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("shutting down matcher gracefully")
			close(shutdown)

			select {
			case <-m.done:
				logger.Info("matcher shutdown completed successfully")
			case <-ctx.Done():
				logger.Warn("matcher shutdown timed out")
			}

			m.client.Close()

			return nil
		},
	})

	return m, nil
}

// newClient builds a Kafka client dedicated to the matcher. It is a direct
// (group-less) consumer of every partition of the orders topic, always starting
// at the beginning: the order book is a materialised view of the log, so every
// boot replays the whole log and rebuilds every book from scratch. With no
// committed offsets there is nothing to resume from, which keeps matching
// correct across restarts at the cost of replay time. It produces trades on the
// same client.
//
// Because a group-less consumer reads all partitions itself, run a SINGLE
// matcher instance — it is the sole writer for the whole topic.
func newClient(cfg kafka.Config, tele telemetry.Telemetry) (*kgo.Client, error) {
	tracer := kotel.NewTracer(
		kotel.TracerProvider(tele.TraceProvider),
		kotel.TracerPropagator(propagation.TraceContext{}),
	)
	hooks := kotel.NewKotel(kotel.WithTracer(tracer)).Hooks()

	client, err := kgo.NewClient(
		kgo.SeedBrokers(cfg.Seeds...),
		kgo.ConsumeTopics(constant.Topic),
		kgo.ConsumeResetOffset(kgo.NewOffset().AtStart()),
		kgo.WithHooks(hooks...),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create matcher kafka client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), constant.PingTimeout)
	defer cancel()

	if err := client.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping kafka cluster: %w", err)
	}

	return client, nil
}

// run is the sequential matching loop. It owns the engine for its whole life, so
// no locking is needed.
func (m *Matcher) run(ctx context.Context) {
	defer close(m.done)

	m.logger.Info("matcher started, materialising order books from the log")

	for {
		if ctx.Err() != nil {
			m.logger.Info("matcher stopped")

			return
		}

		fetches := m.client.PollFetches(ctx)

		if errs := fetches.Errors(); len(errs) > 0 {
			for _, e := range errs {
				if ctx.Err() != nil {
					return
				}

				m.logger.Error("failed to fetch orders", zap.Error(e.Err), zap.String("topic", e.Topic))
			}

			continue
		}

		fetches.EachRecord(func(record *kgo.Record) {
			m.handle(ctx, record)
		})
	}
}

// handle matches a single order record and turns each resulting fill into a
// trade event and a database row.
func (m *Matcher) handle(ctx context.Context, record *kgo.Record) {
	start := time.Now()

	var order model.Order
	if err := json.Unmarshal(record.Value, &order); err != nil {
		m.logger.Error("failed to decode order", zap.Error(err), zap.ByteString("record", record.Value))

		return
	}

	trades, resting, err := m.engine.Submit(order)
	if err != nil {
		m.logger.Warn("rejected invalid order", zap.Error(err), zap.String("order", order.ID))

		return
	}

	m.metric.OrdersMatched.Inc()

	for _, trade := range trades {
		// Stamp the trade deterministically so replaying the log reproduces the
		// exact same trade: the id is derived from the unique (buy, sell) order
		// pair and the time comes from the order record, not the wall clock.
		trade.ID = tradeID(trade)
		trade.CreatedAt = record.Timestamp

		m.emit(ctx, trade)
	}

	m.metric.MatchLatency.Observe(time.Since(start).Seconds())
	m.logBook(order, trades, resting)
}

// emit publishes a trade to the trades topic and persists it. A failure on
// either path is logged but does not stop matching: the orders log remains the
// source of truth and the book stays consistent.
func (m *Matcher) emit(ctx context.Context, trade model.Trade) {
	data, err := json.Marshal(trade)
	if err != nil {
		m.logger.Error("failed to encode trade", zap.Error(err), zap.String("trade", trade.ID))

		return
	}

	// nolint: exhaustruct
	record := &kgo.Record{
		Topic:     constant.TradesTopic,
		Key:       []byte(trade.Symbol),
		Value:     data,
		Timestamp: trade.CreatedAt,
	}

	if err := m.client.ProduceSync(ctx, record).FirstErr(); err != nil {
		m.logger.Error("failed to publish trade", zap.Error(err), zap.String("trade", trade.ID))
	}

	if err := m.persist(ctx, trade); err != nil {
		m.logger.Error("failed to persist trade", zap.Error(err), zap.String("trade", trade.ID))
	}

	m.metric.TradesProduced.Inc()
}

// tradeNamespace seeds deterministic trade UUIDs (UUIDv5).
const tradeNamespace = "0b6f3a2e-4d1c-4f8a-9b2e-1f3c5d7e9a11"

// tradeID derives a stable UUID from the unique (buy, sell) order pair. A taker
// crosses any given maker at most once, so the pair identifies the trade — and
// because matching is deterministic, replaying the log yields the same id, which
// makes ON CONFLICT DO NOTHING a true dedupe.
func tradeID(t model.Trade) string {
	return uuid.NewSHA1(uuid.MustParse(tradeNamespace), []byte(t.BuyOrderID+":"+t.SellOrderID)).String()
}

// persist writes a trade to PostgreSQL. The id is deterministic, so re-inserting
// the same trade (e.g. on a replay) is a no-op.
func (m *Matcher) persist(ctx context.Context, trade model.Trade) error {
	_, err := m.db.Exec(
		ctx,
		`INSERT INTO trades (id, symbol, price, quantity, buy_order_id, sell_order_id, taker_side, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8) ON CONFLICT (id) DO NOTHING`,
		trade.ID, trade.Symbol, trade.Price, trade.Quantity,
		trade.BuyOrderID, trade.SellOrderID, string(trade.TakerSide), trade.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting trade failed: %w", err)
	}

	return nil
}

// logBook records the outcome of an order plus the top of the book it touched.
func (m *Matcher) logBook(order model.Order, trades []model.Trade, resting uint64) {
	book, _ := m.engine.Book(order.Symbol())

	m.logger.Info("order matched",
		zap.String("order", order.ID),
		zap.String("symbol", order.Symbol()),
		zap.String("side", string(order.Side)),
		zap.Int("trades", len(trades)),
		zap.Uint64("resting", resting),
		zap.Any("book", book.Snapshot(snapshotDepth)),
	)
}
