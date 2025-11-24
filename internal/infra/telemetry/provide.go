package telemetry

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.uber.org/fx"
)

func setupTraceExporter(cfg Config) trace.SpanExporter {
	if !cfg.Trace.Enabled {
		exporter, err := stdouttrace.New(
			stdouttrace.WithPrettyPrint(),
		)
		if err != nil {
			log.Fatalf("failed to initialize export pipeline for traces (stdout): %v", err)
		}

		return exporter
	}

	exporter, err := otlptracegrpc.New(
		context.Background(),
		otlptracegrpc.WithEndpoint(cfg.Trace.Endpoint), otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		log.Fatalf("failed to initialize export pipeline for traces (otlp with grpc): %v", err)
	}

	return exporter
}

func setupMeterExporter(reg *prometheus.Registry, cfg Config) *http.Server {
	if !cfg.Meter.Enabled {
		return nil
	}

	reg.MustRegister(collectors.NewGoCollector())

	srv := http.NewServeMux()
	srv.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg})) // nolint: exhaustruct

	return &http.Server{
		Addr:                         cfg.Meter.Address,
		Handler:                      srv,
		DisableGeneralOptionsHandler: false,
		TLSConfig:                    nil,
		ReadTimeout:                  time.Second,
		ReadHeaderTimeout:            time.Second,
		WriteTimeout:                 time.Second,
		IdleTimeout:                  time.Second,
		MaxHeaderBytes:               0,
		TLSNextProto:                 nil,
		ConnState:                    nil,
		ErrorLog:                     nil,
		BaseContext:                  nil,
		ConnContext:                  nil,
		HTTP2:                        nil,
		Protocols:                    nil,
	}
}

func Provide(lc fx.Lifecycle, cfg Config) Telemetery {
	reg := prometheus.NewRegistry()
	srv := setupMeterExporter(reg, cfg)

	exporter := setupTraceExporter(cfg)

	res, err := resource.Merge(
		resource.Default(),
		resource.NewSchemaless(
			semconv.ServiceNamespaceKey.String(cfg.Namespace),
			semconv.ServiceNameKey.String(cfg.ServiceName),
		),
	)
	if err != nil {
		panic(err)
	}

	bsp := trace.NewBatchSpanProcessor(exporter)
	tp := trace.NewTracerProvider(trace.WithSpanProcessor(bsp), trace.WithResource(res))

	otel.SetTracerProvider(tp)

	tel := Telemetery{
		ServiceName:   cfg.ServiceName,
		Namespace:     cfg.Namespace,
		metricSrv:     srv,
		TraceProvider: tp,
		MeterRegistry: reg,
	}

	lc.Append(
		fx.Hook{
			OnStart: tel.run,
			OnStop:  tel.shutdown,
		},
	)

	return tel
}

func (t Telemetery) run(ctx context.Context) error {
	if t.metricSrv != nil {
		lc := new(net.ListenConfig)

		l, err := lc.Listen(ctx, "tcp", t.metricSrv.Addr)
		if err != nil {
			return fmt.Errorf("metric server listen failed: %w", err)
		}

		go func() {
			err := t.metricSrv.Serve(l)
			if !errors.Is(err, http.ErrServerClosed) {
				log.Fatalf("metric server initiation failed: %v", err)
			}
		}()
	}

	return nil
}

func (t Telemetery) shutdown(ctx context.Context) error {
	err := t.metricSrv.Shutdown(ctx)
	if err != nil {
		return fmt.Errorf("cannot shutdown the metric server %w", err)
	}

	return nil
}
