package kafka

import (
	"context"
	"fmt"

	"github.com/1995parham-teaching/redpanda101/internal/infra/constant"
	"github.com/1995parham-teaching/redpanda101/internal/infra/telemetry"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/plugin/kotel"
	"go.opentelemetry.io/otel/propagation"
	"go.uber.org/fx"
)

func Provide(lc fx.Lifecycle, tel telemetry.Telemetery, cfg Config) (*kgo.Client, error) {
	tracer := kotel.NewTracer(
		kotel.TracerProvider(tel.TraceProvider),
		kotel.TracerPropagator(propagation.TraceContext{}),
	)

	meter := kotel.NewMeter(
		kotel.MeterProvider(tel.MeterProvider),
	)

	kotel := kotel.NewKotel(
		kotel.WithTracer(tracer),
		kotel.WithMeter(meter),
	)

	client, err := kgo.NewClient(
		kgo.SeedBrokers(cfg.Seeds...),
		kgo.ConsumerGroup(cfg.ConsumerGroup),
		kgo.WithHooks(kotel.Hooks()...),
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
