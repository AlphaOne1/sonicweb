// SPDX-FileCopyrightText: 2026 The SonicWeb contributors.
// SPDX-License-Identifier: MPL-2.0

package instrumentation

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	otelprom "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	otellog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.39.0"
)

// OTLPProtocolGRPC specifies the protocol string for gRPC in OpenTelemetry exporters.
const OTLPProtocolGRPC = "grpc"

// OTLPProtocolHTTP specifies the protocol string for HTTP in OpenTelemetry exporters.
const OTLPProtocolHTTP = "http"

// OTLPExporterConsole represents the constant name for the console-based OpenTelemetry exporter.
const OTLPExporterConsole = "console"

// OTLPExporterNone represents the "none" value for disabling OTLP exporters in telemetry configurations.
const OTLPExporterNone = "none"

// OTLPExporterOTLP represents the constant name for the OTLP-based OpenTelemetry exporter.
const OTLPExporterOTLP = "otlp"

// OTLPExporterPrometheus represents the constant name for the Prometheus-based OpenTelemetry exporter.
const OTLPExporterPrometheus = "prometheus"

// MinShutdownTimeout defines the minimum timeout duration for cleanup operations during a shutdown sequence.
const MinShutdownTimeout = 5 * time.Second

// ErrUnsupportedOTLPProtocol indicates that the provided OTLP protocol is not supported.
var ErrUnsupportedOTLPProtocol = errors.New("unsupported protocol")

// SetupOTelSDK bootstraps the OpenTelemetry pipeline.
// If it does not return an error, make sure to call shutdown for proper cleanup.
//
//nolint:funlen // cannot go around the steps, splitting will confuse more than longer function
func SetupOTelSDK(
	ctx context.Context,
	serverName string,
	buildInfoTag string,
	log *slog.Logger) (func(context.Context) error, http.Handler, *slog.Logger, error) {

	if log == nil {
		// get a "do nothing" logger if none is set
		log = slog.New(slog.DiscardHandler)
	}

	var shutdownFuncs []func(context.Context) error
	var err error

	// shutdown calls cleanup functions registered via shutdownFuncs.
	// The errors from the calls are joined.
	// Each registered cleanup will be invoked once.
	shutdown := func(ctx context.Context) error {
		var err error
		for _, fn := range slices.Backward(shutdownFuncs) {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil

		return err
	}

	// handleErr calls shutdown for cleanup and makes sure that all errors are returned.
	handleErr := func(inErr error) {
		shutdownCtx := context.WithoutCancel(ctx)
		var cancel context.CancelFunc = func() {}

		if t, hasDeadline := ctx.Deadline(); ctx.Err() != nil ||
			hasDeadline && time.Until(t) < MinShutdownTimeout {
			shutdownCtx, cancel = context.WithTimeout(context.WithoutCancel(ctx), MinShutdownTimeout)
		}

		defer cancel()

		// Joining with a potentially existing previous err, just to make sure to not miss anything.
		// Normal handling requires that return is called immediately after calling this handling function.
		err = errors.Join(err, inErr, shutdown(shutdownCtx))
	}

	// Set up a propagator.
	prop := newPropagator(log)
	otel.SetTextMapPropagator(prop)

	// Get the common resources to use
	res, rscErr := newResource(ctx, serverName, buildInfoTag)
	if rscErr != nil {
		handleErr(rscErr)
		return shutdown, nil, nil, err
	}

	// Set up a trace provider.
	tracerProvider, trcErr := newTracerProvider(ctx, res, log)
	if trcErr != nil {
		handleErr(trcErr)
		return shutdown, nil, nil, err
	}

	if tracerProvider != nil {
		shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
		otel.SetTracerProvider(tracerProvider)
	}

	// Set up meter provider.
	meterProvider, metricHandler, mtrErr := newMeterProvider(ctx, res, log)
	if mtrErr != nil {
		handleErr(mtrErr)
		return shutdown, nil, nil, err
	}

	if meterProvider != nil {
		shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)
		otel.SetMeterProvider(meterProvider)
	}

	// Set up a logger provider.
	loggerProvider, logErr := newLoggerProvider(ctx, res, log)
	if logErr != nil {
		handleErr(logErr)
		return shutdown, nil, nil, err
	}

	var resultLogger *slog.Logger

	if loggerProvider != nil {
		shutdownFuncs = append(shutdownFuncs, loggerProvider.Shutdown)
		global.SetLoggerProvider(loggerProvider)

		resultLogger = slog.New(otelslog.NewHandler("otel", otelslog.WithLoggerProvider(loggerProvider)))
	}

	return shutdown, metricHandler, resultLogger, err
}

// newPropagator creates a composite propagator based on the environment variable OTEL_PROPAGATORS. Supported
// propagators depend on the Go OpenTelemetry implementation and are currently limited to TraceContext and Baggage.
// If none is given, defaults to TraceContext and Baggage together.
//
// Environment variable processing:
// manual:
//   - OTEL_PROPAGATORS
//
//nolint:ireturn // the result is an interface, no choice here
func newPropagator(log *slog.Logger) propagation.TextMapPropagator {
	propagators := make([]propagation.TextMapPropagator, 0, 2)

	if envPropagators := os.Getenv("OTEL_PROPAGATORS"); envPropagators != "" {
		for p := range strings.SplitSeq(envPropagators, ",") {
			p = strings.TrimSpace(p)

			switch p {
			case "baggage":
				propagators = append(propagators, propagation.Baggage{})
			case "tracecontext":
				propagators = append(propagators, propagation.TraceContext{})
			default:
				log.Warn("unsupported propagator in OTEL_PROPAGATORS", slog.String("name", p))
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

// newResource configures a resource to be used by the telemetry providers.
// Attributes configured in the environment variable `OTEL_RESOURCE_ATTRIBUTES` are added.
// It should contain a comma-separated list of key-value-pairs the form `key0=val0,key1=val1,...`.
func newResource(ctx context.Context, serverName, buildInfoTag string) (*resource.Resource, error) {
	res, err := resource.New(
		ctx,
		resource.WithSchemaURL(semconv.SchemaURL),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serverName),
			semconv.ServiceVersionKey.String(buildInfoTag),
		),
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
		resource.WithProcess(),
		resource.WithContainer(),
		resource.WithOS(),
		resource.WithHost(),
	)

	if err != nil {
		return nil, fmt.Errorf("could not create resource: %w", err)
	}

	return res, nil
}

// newTraceExporter initializes a trace.SpanExporter based on the provided exporter name and protocol.
//
//nolint:ireturn
func newTraceExporter(ctx context.Context, name, protocol string, log *slog.Logger) (trace.SpanExporter, error) {
	var exp trace.SpanExporter
	var err error

	switch name {
	case OTLPExporterOTLP:
		switch protocol {
		case OTLPProtocolGRPC:
			exp, err = otlptracegrpc.New(ctx)
		case OTLPProtocolHTTP:
			exp, err = otlptracehttp.New(ctx)
		default:
			err = ErrUnsupportedOTLPProtocol
		}
	case OTLPExporterConsole:
		exp, err = stdouttrace.New(stdouttrace.WithPrettyPrint())
	default:
		log.Warn("unsupported trace exporter", slog.String("name", name))
	}

	if err != nil {
		return nil, fmt.Errorf("error creating trace exporter: %w", err)
	}

	return exp, nil
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
func newTracerProvider(ctx context.Context, res *resource.Resource, log *slog.Logger) (*trace.TracerProvider, error) {
	traceExporters := make([]trace.SpanExporter, 0, 1)

	envExporters := os.Getenv("OTEL_TRACES_EXPORTER")

	if envExporters == OTLPExporterNone || envExporters == "" {
		return nil, nil //nolint:nilnil // it is completely valid to have no provider set
	}

	protocol := os.Getenv("OTEL_EXPORTER_OTLP_PROTOCOL")

	if specializedProtocol := os.Getenv("OTEL_EXPORTER_OTLP_TRACES_PROTOCOL"); specializedProtocol != "" {
		protocol = specializedProtocol
	}

	for exporter := range strings.SplitSeq(envExporters, ",") {
		exp, err := newTraceExporter(ctx, strings.TrimSpace(exporter), protocol, log)

		if err != nil {
			return nil, fmt.Errorf("could not instantiate trace exporter %v with protocol %v: %w",
				exporter, protocol, err)
		}

		if exp != nil {
			traceExporters = append(traceExporters, exp)
		}
	}

	tracerProviderOptions := make([]trace.TracerProviderOption, 0, len(traceExporters)+1)

	for _, t := range traceExporters {
		tracerProviderOptions = append(tracerProviderOptions, trace.WithBatcher(t))
	}

	if res != nil {
		tracerProviderOptions = append(tracerProviderOptions, trace.WithResource(res))
	}

	tracerProvider := trace.NewTracerProvider(tracerProviderOptions...)

	return tracerProvider, nil
}

// newMeterReader initializes and returns a metric.Exporter, metric.Reader,
// and optional http.Handler based on the given name and protocol.
//
//nolint:ireturn
func newMeterReader(
	ctx context.Context,
	name string,
	protocol string,
	log *slog.Logger) (metric.Exporter, metric.Reader, http.Handler, error) {

	var reader metric.Reader
	var exp metric.Exporter
	var metricHandler http.Handler
	var err error

	switch name {
	case OTLPExporterOTLP:
		switch protocol {
		case OTLPProtocolGRPC:
			exp, err = otlpmetricgrpc.New(ctx)
		case OTLPProtocolHTTP:
			exp, err = otlpmetrichttp.New(ctx)
		default:
			err = ErrUnsupportedOTLPProtocol
		}
	case OTLPExporterPrometheus:
		reg := prometheus.NewRegistry()                                  // create explicit registry
		reader, err = otelprom.New(otelprom.WithRegisterer(reg))         // create reader that writes to that registry
		metricHandler = promhttp.HandlerFor(reg, promhttp.HandlerOpts{}) // create handlers reading exactly that registry
	case OTLPExporterConsole:
		exp, err = stdoutmetric.New()
	default:
		log.Warn("unsupported metric exporter", slog.String("name", name))
	}

	if err != nil {
		return nil, nil, nil, fmt.Errorf("error creating meter reader: %w", err)
	}

	return exp, reader, metricHandler, nil
}

// newMeterProvider creates a new meter provider based on the environment variables.
// For the environment variable `OTEL_METRICS_EXPORTER` it supports the values `otlp`,
// `prometheus`, `console` and `none`, with `none` being the default.
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
func newMeterProvider(
	ctx context.Context,
	res *resource.Resource,
	log *slog.Logger) (*metric.MeterProvider, http.Handler, error) {

	metricReaders := make([]metric.Reader, 0, 1)

	envExporters := os.Getenv("OTEL_METRICS_EXPORTER")

	if envExporters == OTLPExporterNone || envExporters == "" {
		return nil, nil, nil
	}

	protocol := os.Getenv("OTEL_EXPORTER_OTLP_PROTOCOL")

	if specializedProtocol := os.Getenv("OTEL_EXPORTER_OTLP_METRICS_PROTOCOL"); specializedProtocol != "" {
		protocol = specializedProtocol
	}

	var metricHandler http.Handler

	for exporter := range strings.SplitSeq(envExporters, ",") {
		exp, reader, tmpHandler, err := newMeterReader(ctx, strings.TrimSpace(exporter), protocol, log)

		if err != nil {
			return nil, nil, fmt.Errorf("could not instantiate metrics exporter %v with protocol %v: %w",
				exporter, protocol, err)
		}

		if tmpHandler != nil {
			metricHandler = tmpHandler
		}

		if exp != nil {
			reader = metric.NewPeriodicReader(exp)
		}

		if reader != nil {
			metricReaders = append(metricReaders, reader)
		}
	}

	meterProviderOptions := make([]metric.Option, 0, len(metricReaders)+1)

	for _, r := range metricReaders {
		meterProviderOptions = append(meterProviderOptions, metric.WithReader(r))
	}

	if res != nil {
		meterProviderOptions = append(meterProviderOptions, metric.WithResource(res))
	}

	meterProvider := metric.NewMeterProvider(meterProviderOptions...)

	return meterProvider, metricHandler, nil
}

// newLoggerExporter creates a log exporter based on the provided name and protocol or returns an error if unsupported.
//
//nolint:ireturn
func newLoggerExporter(ctx context.Context, name, protocol string, log *slog.Logger) (otellog.Exporter, error) {
	var exp otellog.Exporter
	var err error

	switch name {
	case OTLPExporterOTLP:
		switch protocol {
		case OTLPProtocolGRPC:
			exp, err = otlploggrpc.New(ctx)
		case OTLPProtocolHTTP:
			exp, err = otlploghttp.New(ctx)
		default:
			err = ErrUnsupportedOTLPProtocol
		}
	case OTLPExporterConsole:
		exp, err = stdoutlog.New()
	default:
		log.Warn("unsupported log exporter", slog.String("name", name))
	}

	if err != nil {
		return nil, fmt.Errorf("error creating logger exporter: %w", err)
	}

	return exp, nil
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
func newLoggerProvider(ctx context.Context, res *resource.Resource, log *slog.Logger) (*otellog.LoggerProvider, error) {
	logExporters := make([]otellog.Exporter, 0, 1)

	envExporters := os.Getenv("OTEL_LOGS_EXPORTER")

	if envExporters == OTLPExporterNone || envExporters == "" {
		return nil, nil //nolint:nilnil // it is completely valid to have no provider set
	}

	protocol := os.Getenv("OTEL_EXPORTER_OTLP_PROTOCOL")

	if specializedProtocol := os.Getenv("OTEL_EXPORTER_OTLP_LOGS_PROTOCOL"); specializedProtocol != "" {
		protocol = specializedProtocol
	}

	for exporter := range strings.SplitSeq(envExporters, ",") {
		exp, err := newLoggerExporter(ctx, strings.TrimSpace(exporter), protocol, log)

		if err != nil {
			return nil, fmt.Errorf("could not instantiate log exporter %v with protocol %v: %w",
				exporter, protocol, err)
		}

		if exp != nil {
			logExporters = append(logExporters, exp)
		}
	}

	loggerProviderOptions := make([]otellog.LoggerProviderOption, 0, len(logExporters)+1)

	for _, l := range logExporters {
		loggerProviderOptions = append(loggerProviderOptions, otellog.WithProcessor(otellog.NewBatchProcessor(l)))
	}

	if res != nil {
		loggerProviderOptions = append(loggerProviderOptions, otellog.WithResource(res))
	}

	loggerProvider := otellog.NewLoggerProvider(loggerProviderOptions...)

	return loggerProvider, nil
}
