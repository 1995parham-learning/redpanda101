package consumer

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type Metric struct {
	DatabaseInsertionTime prometheus.Histogram
}

func NewMetric(reg *prometheus.Registry, namespace, serviceName string) *Metric {
	dit := prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace:                       namespace,
		Subsystem:                       serviceName,
		Name:                            "database_insertion_time_seconds",
		Help:                            "time spend for inserting a record into the database in seconds",
		ConstLabels:                     prometheus.Labels{},
		Buckets:                         prometheus.DefBuckets,
		NativeHistogramBucketFactor:     0,
		NativeHistogramZeroThreshold:    0,
		NativeHistogramMaxBucketNumber:  0,
		NativeHistogramMinResetDuration: 0,
		NativeHistogramMaxZeroThreshold: 0,
		NativeHistogramMaxExemplars:     0,
		NativeHistogramExemplarTTL:      0,
	})

	if err := reg.Register(dit); err != nil {
		panic(err)
	}

	return &Metric{
		DatabaseInsertionTime: dit,
	}
}

func (m *Metric) DatabaseInsertionTimeRecord(d time.Duration) {
	m.DatabaseInsertionTime.Observe(d.Seconds())
}
