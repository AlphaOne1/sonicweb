// SPDX-FileCopyrightText: 2025 The SonicWeb contributors.
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
	"strings"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.27.0"
)

// setupOTelSDK bootstraps the OpenTelemetry pipeline.
// If it does not return an error, make sure to call shutdown for proper cleanup.
func setupOTelSDK(ctx context.Context) (func(context.Context) error, error) {
	var shutdownFuncs []func(context.Context) error
	var err error

	// shutdown calls cleanup functions registered via shutdownFuncs.
	// The errors from the calls are joined.
	// Each registered cleanup will be invoked once.
	shutdown := func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil

		return err
	}

	// handleErr calls shutdown for cleanup and makes sure that all errors are returned.
	handleErr := func(inErr error) {
		err = errors.Join(inErr, shutdown(ctx))
	}

	// Set up a propagator.
	prop := newPropagator()
	otel.SetTextMapPropagator(prop)

	// Set up a trace provider.
	tracerProvider, err := newTracerProvider()
	if err != nil {
		handleErr(err)
		return shutdown, err
	}
	shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
	otel.SetTracerProvider(tracerProvider)

	// Set up meter provider.
	meterProvider, err := newMeterProvider()
	if err != nil {
		handleErr(err)
		return shutdown, err
	}
	shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)
	otel.SetMeterProvider(meterProvider)

	// Set up a logger provider.
	loggerProvider, err := newLoggerProvider()
	if err != nil {
		handleErr(err)
		return shutdown, err
	}
	shutdownFuncs = append(shutdownFuncs, loggerProvider.Shutdown)
	global.SetLoggerProvider(loggerProvider)

	return shutdown, err
}

// newPropagator creates a composite propagator based on the environment variable OTEL_PROPAGATORS. Supported
// propagators depend on the Go OpenTelemetry implementation and are currently limited to TraceContext and Baggage.
// If none is given, defaults to TraceContext and Baggage together.
//
// Environment variable processing:
// manual:
//   - OTEL_PROPAGATORS
func newPropagator() propagation.TextMapPropagator {
	propagators := make([]propagation.TextMapPropagator, 0, 2)

	if envPropagators := os.Getenv("OTEL_PROPAGATORS"); envPropagators != "" {
		for _, p := range strings.Split(envPropagators, ",") {
			switch p {
			case "baggage":
				propagators = append(propagators, propagation.Baggage{})
			case "tracecontext":
				propagators = append(propagators, propagation.TraceContext{})
			default:
				slog.Warn("unsupported propagator in OTEL_PROPAGATORS", slog.String("name", p))
			}
		}
	}

	if len(propagators) == 0 {
		propagators = append(propagators,
			propagation.TraceContext{},
			propagation.Baggage{})
	}

	return propagation.NewCompositeTextMapPropagator(propagators...)
}

// newTracerProvider creates a new trace provider based on the environment variables. For the environment variable
// `OTEL_TRACES_EXPORTER` it supports the values `otlp`, `console` and `none`, with `none` being the default.
//
// Environment variable processing:
// manual:
//   - OTEL_EXPORTER_OTLP_PROTOCOL,           OTEL_EXPORTER_OTLP_TRACES_PROTOCOL
//   - OTEL_TRACES_EXPORTER
//
// automatic:
//   - OTEL_EXPORTER_OTLP_ENDPOINT,           OTEL_EXPORTER_OTLP_TRACES_ENDPOINT
//   - OTEL_EXPORTER_OTLP_HEADERS,            OTEL_EXPORTER_OTLP_TRACES_HEADERS
//   - OTEL_EXPORTER_OTLP_TIMEOUT,            OTEL_EXPORTER_OTLP_TRACES_TIMEOUT
//   - OTEL_EXPORTER_OTLP_COMPRESSION,        OTEL_EXPORTER_OTLP_TRACES_COMPRESSION
//   - OTEL_EXPORTER_OTLP_CERTIFICATE,        OTEL_EXPORTER_OTLP_TRACES_CERTIFICATE
//   - OTEL_EXPORTER_OTLP_CLIENT_CERTIFICATE, OTEL_EXPORTER_OTLP_TRACES_CLIENT_CERTIFICATE
//   - OTEL_EXPORTER_OTLP_CLIENT_KEY,         OTEL_EXPORTER_OTLP_TRACES_CLIENT_KEY
//   - OTEL_TRACES_SAMPLER
//   - OTEL_TRACES_SAMPLER_ARG
func newTracerProvider() (*trace.TracerProvider, error) {
	traceExporters := make([]trace.SpanExporter, 0, 1)

	envExporters := os.Getenv("OTEL_TRACES_EXPORTER")

	if envExporters == "none" || envExporters == "" {
		return nil, nil
	}

	protocol := os.Getenv("OTEL_EXPORTER_OTLP_PROTOCOL")

	if specializedProtocol := os.Getenv("OTEL_EXPORTER_OTLP_TRACES_PROTOCOL"); specializedProtocol != "" {
		protocol = specializedProtocol
	}

	for _, exporter := range strings.Split(envExporters, ",") {
		var se trace.SpanExporter
		var err error

		switch exporter {
		case "otlp":
			switch protocol {
			case "grpc":
				se, err = otlptracegrpc.New(context.TODO())
			case "http":
				se, err = otlptracehttp.New(context.TODO())
			default:
				err = fmt.Errorf("unsupported protocol %q for exporter %q", protocol, exporter)
			}
		case "console":
			se, err = stdouttrace.New(stdouttrace.WithPrettyPrint())
		}

		if err != nil {
			return nil, err
		}

		traceExporters = append(traceExporters, se)
	}

	tracerProviderOptions := make([]trace.TracerProviderOption, 0, len(traceExporters))

	for _, t := range traceExporters {
		tracerProviderOptions = append(tracerProviderOptions, trace.WithBatcher(t))
	}

	tracerProvider := trace.NewTracerProvider(
		tracerProviderOptions...,
	)

	return tracerProvider, nil
}

// newMeterProvider creates a new meter provider based on the environment variables. For the environment variable
// `OTEL_METRICS_EXPORTER` it supports the values `otlp`, `console` and `none`, with `none` being the default.
//
// Environment variable processing:
// manual:
//   - OTEL_EXPORTER_OTLP_PROTOCOL,           OTEL_EXPORTER_OTLP_METRICS_PROTOCOL
//   - OTEL_METRICS_EXPORTER
//
// automatic:
//   - OTEL_EXPORTER_OTLP_ENDPOINT,           OTEL_EXPORTER_OTLP_METRICS_ENDPOINT
//   - OTEL_EXPORTER_OTLP_HEADERS,            OTEL_EXPORTER_OTLP_METRICS_HEADERS
//   - OTEL_EXPORTER_OTLP_TIMEOUT,            OTEL_EXPORTER_OTLP_METRICS_TIMEOUT
//   - OTEL_EXPORTER_OTLP_COMPRESSION,        OTEL_EXPORTER_OTLP_METRICS_COMPRESSION
//   - OTEL_EXPORTER_OTLP_CERTIFICATE,        OTEL_EXPORTER_OTLP_METRICS_CERTIFICATE
//   - OTEL_EXPORTER_OTLP_CLIENT_CERTIFICATE, OTEL_EXPORTER_OTLP_METRICS_CLIENT_CERTIFICATE
//   - OTEL_EXPORTER_OTLP_CLIENT_KEY,         OTEL_EXPORTER_OTLP_METRICS_CLIENT_KEY
//   - OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE
//   - OTEL_EXPORTER_OTLP_METRICS_DEFAULT_HISTOGRAM_AGGREGATION
//   - OTEL_METRIC_EXPORT_INTERVAL
//   - OTEL_METRIC_EXPORT_TIMEOUT
func newMeterProvider() (*metric.MeterProvider, error) {
	metricExporters := make([]metric.Exporter, 0, 1)

	envExporters := os.Getenv("OTEL_METRICS_EXPORTER")

	if envExporters == "none" || envExporters == "" {
		return nil, nil
	}

	protocol := os.Getenv("OTEL_EXPORTER_OTLP_PROTOCOL")

	if specializedProtocol := os.Getenv("OTEL_EXPORTER_OTLP_METRIC_PROTOCOL"); specializedProtocol != "" {
		protocol = specializedProtocol
	}

	for _, exporter := range strings.Split(envExporters, ",") {
		var me metric.Exporter
		var err error

		switch exporter {
		case "otlp":
			switch protocol {
			case "grpc":
				me, err = otlpmetricgrpc.New(context.TODO())
			case "http":
				me, err = otlpmetrichttp.New(context.TODO())
			default:
				err = fmt.Errorf("unsupported protocol %q for exporter %q", protocol, exporter)
			}
		case "console":
			me, err = stdoutmetric.New()
		}

		if err != nil {
			return nil, err
		}

		metricExporters = append(metricExporters, me)
	}

	meterProviderOptions := make([]metric.Option, 0, len(metricExporters))

	for _, m := range metricExporters {
		meterProviderOptions = append(meterProviderOptions, metric.WithReader(metric.NewPeriodicReader(m)))
	}

	meterProvider := metric.NewMeterProvider(
		meterProviderOptions...,
	)

	return meterProvider, nil
}

// newLoggerProvider creates a new logger provider based on the environment variables. For the environment variable
// `OTEL_LOGS_EXPORTER` it supports the values `otlp`, `console` and `none`, with `none` being the default.
//
// Environment variable processing:
// manual:
//   - OTEL_EXPORTER_OTLP_PROTOCOL,           OTEL_EXPORTER_OTLP_LOGS_PROTOCOL
//   - OTEL_LOGS_EXPORTER
//
// automatic:
//   - OTEL_EXPORTER_OTLP_ENDPOINT,           OTEL_EXPORTER_OTLP_LOGS_ENDPOINT
//   - OTEL_EXPORTER_OTLP_HEADERS,            OTEL_EXPORTER_OTLP_LOGS_HEADERS
//   - OTEL_EXPORTER_OTLP_TIMEOUT,            OTEL_EXPORTER_OTLP_LOGS_TIMEOUT
//   - OTEL_EXPORTER_OTLP_COMPRESSION,        OTEL_EXPORTER_OTLP_LOGS_COMPRESSION
//   - OTEL_EXPORTER_OTLP_CERTIFICATE,        OTEL_EXPORTER_OTLP_LOGS_CERTIFICATE
//   - OTEL_EXPORTER_OTLP_CLIENT_CERTIFICATE, OTEL_EXPORTER_OTLP_LOGS_CLIENT_CERTIFICATE
//   - OTEL_EXPORTER_OTLP_CLIENT_KEY,         OTEL_EXPORTER_OTLP_LOGS_CLIENT_KEY
//   - OTEL_BLRP_SCHEDULE_DELAY
//   - OTEL_BLRP_MAX_EXPORT_BATCH_SIZE
//   - OTEL_BLRP_EXPORT_TIMEOUT
//   - OTEL_BLRP_MAX_QUEUE_SIZE
//   - OTEL_LOGRECORD_ATTRIBUTE_COUNT_LIMIT
//   - OTEL_LOGRECORD_ATTRIBUTE_VALUE_LENGTH_LIMIT
func newLoggerProvider() (*log.LoggerProvider, error) {
	logExporters := make([]log.Exporter, 0, 1)

	envExporters := os.Getenv("OTEL_LOGS_EXPORTER")

	if envExporters == "none" || envExporters == "" {
		return nil, nil
	}

	protocol := os.Getenv("OTEL_EXPORTER_OTLP_PROTOCOL")

	if specializedProtocol := os.Getenv("OTEL_EXPORTER_OTLP_LOGS_PROTOCOL"); specializedProtocol != "" {
		protocol = specializedProtocol
	}

	for _, exporter := range strings.Split(envExporters, ",") {
		var le log.Exporter
		var err error

		switch exporter {
		case "otlp":
			switch protocol {
			case "grpc":
				le, err = otlploggrpc.New(context.TODO())
			case "http":
				le, err = otlploghttp.New(context.TODO())
			default:
				err = fmt.Errorf("unsupported protocol %q for exporter %q", protocol, exporter)
			}
		case "console":
			le, err = stdoutlog.New()
		}

		if err != nil {
			return nil, err
		}

		logExporters = append(logExporters, le)
	}

	loggerProviderOptions := make([]log.LoggerProviderOption, 0, len(logExporters))

	for _, l := range logExporters {
		loggerProviderOptions = append(loggerProviderOptions, log.WithProcessor(log.NewBatchProcessor(l)))
	}

	loggerProvider := log.NewLoggerProvider(
		loggerProviderOptions...,
	)

	return loggerProvider, nil
}

// initTracer initializes the tracing functionality. If used, the tracing
// backend will report tracing data to the provided tracing endpoint.
func initTracer(endpoint string) (*trace.TracerProvider, error) {
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

	// For the demonstration, use trace.AlwaysSample sampler to sample all traces.
	// In a production application, use trace.ProbabilitySampler with a desired probability.
	tracerProvider := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(ServerName))),
		trace.WithSampler(trace.AlwaysSample()),
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
