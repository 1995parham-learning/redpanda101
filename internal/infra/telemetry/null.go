package telemetry

import (
	"github.com/prometheus/client_golang/prometheus"
	tnoop "go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/fx"
)

func ProvideNull(_ fx.Lifecycle) Telemetery {
	tel := Telemetery{
		ServiceName:   "",
		Namespace:     "",
		metricSrv:     nil,
		TraceProvider: tnoop.NewTracerProvider(),
		MeterRegistry: prometheus.NewPedanticRegistry(),
	}

	return tel
}
