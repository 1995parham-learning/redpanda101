package consumer

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/metric"
)

type Metric struct {
	DatabaseInsertionTime metric.Float64Histogram
}

func NewMetric(meter metric.Meter) *Metric {
	dit, err := meter.Float64Histogram("database_insertion_time_seconds")
	if err != nil {
		panic(err)
	}

	return &Metric{
		DatabaseInsertionTime: dit,
	}
}

func (m *Metric) DatabaseInsertionTimeRecord(ctx context.Context, d time.Duration) {
	m.DatabaseInsertionTime.Record(ctx, d.Seconds())
}
