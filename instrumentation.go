// Copyright the SonicWeb contributors.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/http/pprof"
	"os"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/prometheus"

	// "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.27.0"
)

// initTracer initializes the tracing functionality. If used, the tracing
// backend will report tracing data to the provided tracing endpoint.
func initTracer(endpoint string) (*sdktrace.TracerProvider, error) {
	// // Create stdout exporter to be able to retrieve
	// // the collected spans.

	// exporter, err := stdouttrace.New(stdout.WithPrettyPrint())

	exporter, err := otlptracehttp.New(context.Background(),
		otlptracehttp.WithCompression(otlptracehttp.GzipCompression),
		otlptracehttp.WithEndpoint(endpoint),
	)
	if err != nil {
		return nil, fmt.Errorf("could not get telemetry exporter: %w", err)
	}

	// For the demonstration, use sdktrace.AlwaysSample sampler to sample all traces.
	// In a production application, use sdktrace.ProbabilitySampler with a desired probability.
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(ServerName))),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tracerProvider, nil
}

// setupMetricsInstrumentation sets up the telemetry functionality.
// If either telemetry or pprof are enabled, serveMetrics is called subsequently to
// open a telemetry and/or profiling providing server port.
func setupMetricsInstrumentation(
	instrumentAddress *string,
	instrumentPort *string,
	enableTelemetry bool,
	enablePprof bool) {

	if !enableTelemetry && !enablePprof {
		return
	}

	if enableTelemetry {
		metricsExporter, metricsExporterErr := prometheus.New()

		if metricsExporterErr != nil {
			slog.Error("could not create prometheus exporter", slog.String("error", metricsExporterErr.Error()))
			os.Exit(1)
		}

		baseResource, baseResourceErr := resource.Merge(resource.Default(),
			resource.NewSchemaless(
				attribute.KeyValue{
					Key:   "service.name",
					Value: attribute.StringValue("sonic-web"),
				}))

		if baseResourceErr != nil {
			slog.Error("could not create base resource", slog.String("error", baseResourceErr.Error()))
		}

		meterProvider := metric.NewMeterProvider(
			metric.WithReader(metricsExporter),
			metric.WithResource(baseResource))

		otel.SetMeterProvider(meterProvider)
	}

	go serveMetrics(*instrumentAddress, *instrumentPort, enableTelemetry, enablePprof)
}

// serveMetrics sets up the metrics functionality. It opens a separate port, and metrics collectors can fetch their
// data from there. For profiling purposes, if enabled, also the pprof endpoints are added on the same port.
func serveMetrics(address string, port string, enableTelemetry, enablePprof bool) {
	if !enableTelemetry && !enablePprof {
		return
	}

	listenAddress := net.JoinHostPort(address, port)

	mux := http.NewServeMux()

	if enablePprof {
		slog.Info("serving pprof", slog.String("address", listenAddress+"/debug/pprof"))
		mux.HandleFunc("GET /debug/pprof/", pprof.Index)
		mux.HandleFunc("GET /debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("GET /debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("GET /debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("GET /debug/pprof/trace", pprof.Trace)
	} else {
		slog.Info("serving pprof disabled")
	}

	if enableTelemetry {
		slog.Info("serving telemetry", slog.String("address", listenAddress+"/metrics"))
		mux.Handle("GET /metrics", promhttp.Handler())
	} else {
		slog.Info("serving telemetry disabled")
	}

	server := http.Server{
		Addr:              listenAddress,
		Handler:           mux,
		ReadHeaderTimeout: ReadTimeout,
		ReadTimeout:       ReadTimeout,
	}

	defer func() { _ = server.Close() }()

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("error serving metrics", slog.String("error", err.Error()))
			os.Exit(1)
		}
		slog.Info("metrics server stopped to accept new connections")
	}()

	if shutdownErr := waitServerShutdown(&server, "metrics"); shutdownErr != nil {
		slog.Error("error shutting down metrics server", slog.String("error", shutdownErr.Error()))
	}
}
