package producer

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Metric struct {
	MessagesProduced prometheus.Counter
	ProduceLatency   prometheus.Histogram
	ProduceErrors    prometheus.Counter
}

func NewMetric(reg *prometheus.Registry, namespace, serviceName string) *Metric {
	mp := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace:   namespace,
		Subsystem:   serviceName,
		Name:        "messages_produced_total",
		Help:        "total number of messages produced to kafka",
		ConstLabels: prometheus.Labels{},
	})

	err := reg.Register(mp)
	if err != nil {
		panic(err)
	}

	pl := prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace:                       namespace,
		Subsystem:                       serviceName,
		Name:                            "produce_latency_seconds",
		Help:                            "time spent producing a message to kafka in seconds",
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

	err = reg.Register(pl)
	if err != nil {
		panic(err)
	}

	pe := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace:   namespace,
		Subsystem:   serviceName,
		Name:        "produce_errors_total",
		Help:        "total number of errors while producing messages to kafka",
		ConstLabels: prometheus.Labels{},
	})

	err = reg.Register(pe)
	if err != nil {
		panic(err)
	}

	return &Metric{
		MessagesProduced: mp,
		ProduceLatency:   pl,
		ProduceErrors:    pe,
	}
}
