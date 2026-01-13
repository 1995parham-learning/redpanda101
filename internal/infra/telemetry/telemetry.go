package telemetry

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel/trace"
)

type Telemetry struct {
	ServiceName   string
	Namespace     string
	metricSrv     *http.Server
	TraceProvider trace.TracerProvider
	MeterRegistry *prometheus.Registry
}
