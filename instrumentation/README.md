<!-- markdownlint-disable MD013 MD033 MD041 -->
<!-- SPDX-FileCopyrightText: 2026 The SonicWeb contributors.
     SPDX-License-Identifier: MPL-2.0
-->

SonicWeb Instrumentation
========================

This library is derived from the official documentation example for the OpenTelemetry Go SDK. It aims to provide a thin
layer of automatic environment variable-based configuration for OpenTelemetry instrumentation. The OpenTelemetry Go SDK
offers many more customization options but requires a lot of boilerplate code. Use this library if you just want to have
a functioning OpenTelemetry setup without having to dive into the details.

A typical use case is a service in Kubernetes, where the OpenTelemetry configuration is managed by the Kubernetes
operator. The operator can set environment variables to configure the OpenTelemetry SDK. The owner of the service just
provides the means for the operator and does not have further to care about the details of OpenTelemetry.


Installation
------------

To install the *SonicWeb* instrumentation library, you can use the following command:

```sh
go get github.com/AlphaOne1/sonicweb/instrumentation
```

Versions of this library are bound to the semantic versioning of *SonicWeb*. This library is intended for public use but
be aware that breaking changes may occur between minor versions. No breaking changes will be introduced between patch
versions.


Getting Started
---------------

An example that shows how to use this library:

```go
otelShutdown, metricHandler, otelLogger, err := instrumentation.SetupOTelSDK(
	context.Background(),
	"example-service", // service name
	"dev",             // service version / build tag
	slog.Default(),
)

if err != nil { return }

// cleanup internal OpenTelemetry resources on shutdown
defer func() {
    shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer shutdownCancel()
    _ = otelShutdown(shutdownCtx)
}()

// add OpenTelemetry logger to default logger
if otelLogger != nil {
    slog.SetDefault(slog.New(slog.NewMultiHandler(
        slog.Default().Handler(),
        otelLogger.Handler())))
}

// register and expose the metrics handler at the /metrics endpoint of an existing server/mux.
if metricHandler != nil {
    http.Handle("/metrics", metricHandler)
}
```


Environment Variables (OpenTelemetry)
-------------------------------------

This package evaluates a set of OpenTelemetry environment variables directly (manual processing).
In addition, the OpenTelemetry Go SDK and its exporters evaluate further variables automatically.


### Manual processing

The following environment variables are evaluated by the *SonicWeb* instrumentation library.

| Variable                              | Purpose                                                                        | Supported / Expected Values                                                                                 | Default / Behavior                                                                                      |
|---------------------------------------|--------------------------------------------------------------------------------|-------------------------------------------------------------------------------------------------------------|---------------------------------------------------------------------------------------------------------|
| `OTEL_PROPAGATORS`                    | Selects the propagators used for extracting/injecting trace context.           | Comma-separated list. Supported here: `tracecontext`, `baggage`. Unknown values are ignored with a warning. | If unset/empty or only unsupported values are given, defaults to `tracecontext,baggage`.                |
| `OTEL_RESOURCE_ATTRIBUTES`            | Adds resource attributes to all telemetry data (service metadata, etc.).       | Comma-separated `key=value` pairs, e.g. `deployment.environment=prod,team=core`.                            | If unset, only the built-in attributes (service name/version + SDK/process/host/container/OS) are used. |
| `OTEL_TRACES_EXPORTER`                | Enables and selects trace exporter(s).                                         | Comma-separated list. Supported by this package: `otlp`, `console`, `none`.                                 | If unset/empty or `none`: tracing is disabled (no tracer provider).                                     |
| `OTEL_METRICS_EXPORTER`               | Enables and selects metric exporter(s).                                        | Comma-separated list. Supported by this package: `otlp`, `prometheus`, `console`, `none`.                   | If unset/empty or `none`: metrics are disabled (no meter provider).                                     |
| `OTEL_LOGS_EXPORTER`                  | Enables and selects log exporter(s).                                           | Comma-separated list. Supported by this package: `otlp`, `console`, `none`.                                 | If unset/empty or `none`: log exporting is disabled (no logger provider).                               |
| `OTEL_EXPORTER_OTLP_PROTOCOL`         | Selects the OTLP transport protocol (shared fallback for traces/metrics/logs). | `grpc` or `http`.                                                                                           | Used as fallback. A signal-specific protocol variable takes precedence (see below).                     |
| `OTEL_EXPORTER_OTLP_TRACES_PROTOCOL`  | Overrides OTLP protocol for traces.                                            | `grpc` or `http`.                                                                                           | If set, overrides `OTEL_EXPORTER_OTLP_PROTOCOL` for traces.                                             |
| `OTEL_EXPORTER_OTLP_METRICS_PROTOCOL` | Overrides OTLP protocol for metrics.                                           | `grpc` or `http`.                                                                                           | If set, overrides `OTEL_EXPORTER_OTLP_PROTOCOL` for metrics.                                            |
| `OTEL_EXPORTER_OTLP_LOGS_PROTOCOL`    | Overrides OTLP protocol for logs.                                              | `grpc` or `http`.                                                                                           | If set, overrides `OTEL_EXPORTER_OTLP_PROTOCOL` for logs.                                               |


### Automatic processing

The following variables are evaluated by the OpenTelemetry Go SDK and/or OTLP exporters when corresponding exporters are
enabled.


#### Traces (OTLP)

| Variable                                                                                 | Purpose (high-level)                           |
|------------------------------------------------------------------------------------------|------------------------------------------------|
| `OTEL_EXPORTER_OTLP_ENDPOINT` / `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT`                     | OTLP endpoint (generic or trace-specific).     |
| `OTEL_EXPORTER_OTLP_HEADERS` / `OTEL_EXPORTER_OTLP_TRACES_HEADERS`                       | Additional request headers (e.g. auth tokens). |
| `OTEL_EXPORTER_OTLP_TIMEOUT` / `OTEL_EXPORTER_OTLP_TRACES_TIMEOUT`                       | Export timeout.                                |
| `OTEL_EXPORTER_OTLP_COMPRESSION` / `OTEL_EXPORTER_OTLP_TRACES_COMPRESSION`               | Payload compression (e.g. `gzip`).             |
| `OTEL_EXPORTER_OTLP_CERTIFICATE` / `OTEL_EXPORTER_OTLP_TRACES_CERTIFICATE`               | TLS CA certificate for OTLP endpoint.          |
| `OTEL_EXPORTER_OTLP_CLIENT_CERTIFICATE` / `OTEL_EXPORTER_OTLP_TRACES_CLIENT_CERTIFICATE` | mTLS client certificate.                       |
| `OTEL_EXPORTER_OTLP_CLIENT_KEY` / `OTEL_EXPORTER_OTLP_TRACES_CLIENT_KEY`                 | mTLS client key.                               |
| `OTEL_TRACES_SAMPLER`                                                                    | Sampler selection.                             |
| `OTEL_TRACES_SAMPLER_ARG`                                                                | Sampler configuration argument.                |


#### Metrics (OTLP / Prometheus)

| Variable                                                                                  | Purpose (high-level)                        |
|-------------------------------------------------------------------------------------------|---------------------------------------------|
| `OTEL_EXPORTER_OTLP_ENDPOINT` / `OTEL_EXPORTER_OTLP_METRICS_ENDPOINT`                     | OTLP endpoint (generic or metric-specific). |
| `OTEL_EXPORTER_OTLP_HEADERS` / `OTEL_EXPORTER_OTLP_METRICS_HEADERS`                       | Additional request headers.                 |
| `OTEL_EXPORTER_OTLP_TIMEOUT` / `OTEL_EXPORTER_OTLP_METRICS_TIMEOUT`                       | Export timeout.                             |
| `OTEL_EXPORTER_OTLP_COMPRESSION` / `OTEL_EXPORTER_OTLP_METRICS_COMPRESSION`               | Payload compression.                        |
| `OTEL_EXPORTER_OTLP_CERTIFICATE` / `OTEL_EXPORTER_OTLP_METRICS_CERTIFICATE`               | TLS CA certificate.                         |
| `OTEL_EXPORTER_OTLP_CLIENT_CERTIFICATE` / `OTEL_EXPORTER_OTLP_METRICS_CLIENT_CERTIFICATE` | mTLS client certificate.                    |
| `OTEL_EXPORTER_OTLP_CLIENT_KEY` / `OTEL_EXPORTER_OTLP_METRICS_CLIENT_KEY`                 | mTLS client key.                            |
| `OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE`                                       | Metric temporality preference.              |
| `OTEL_EXPORTER_OTLP_METRICS_DEFAULT_HISTOGRAM_AGGREGATION`                                | Histogram aggregation configuration.        |
| `OTEL_METRIC_EXPORT_INTERVAL`                                                             | Periodic export interval.                   |
| `OTEL_METRIC_EXPORT_TIMEOUT`                                                              | Periodic export timeout.                    |


#### Logs (OTLP)

| Variable                                                                               | Purpose (high-level)                         |
|----------------------------------------------------------------------------------------|----------------------------------------------|
| `OTEL_EXPORTER_OTLP_ENDPOINT` / `OTEL_EXPORTER_OTLP_LOGS_ENDPOINT`                     | OTLP endpoint (generic or log-specific).     |
| `OTEL_EXPORTER_OTLP_HEADERS` / `OTEL_EXPORTER_OTLP_LOGS_HEADERS`                       | Additional request headers.                  |
| `OTEL_EXPORTER_OTLP_TIMEOUT` / `OTEL_EXPORTER_OTLP_LOGS_TIMEOUT`                       | Export timeout.                              |
| `OTEL_EXPORTER_OTLP_COMPRESSION` / `OTEL_EXPORTER_OTLP_LOGS_COMPRESSION`               | Payload compression.                         |
| `OTEL_EXPORTER_OTLP_CERTIFICATE` / `OTEL_EXPORTER_OTLP_LOGS_CERTIFICATE`               | TLS CA certificate.                          |
| `OTEL_EXPORTER_OTLP_CLIENT_CERTIFICATE` / `OTEL_EXPORTER_OTLP_LOGS_CLIENT_CERTIFICATE` | mTLS client certificate.                     |
| `OTEL_EXPORTER_OTLP_CLIENT_KEY` / `OTEL_EXPORTER_OTLP_LOGS_CLIENT_KEY`                 | mTLS client key.                             |
| `OTEL_BLRP_SCHEDULE_DELAY`                                                             | Batch log record processor scheduling delay. |
| `OTEL_BLRP_MAX_EXPORT_BATCH_SIZE`                                                      | Batch size limit per export.                 |
| `OTEL_BLRP_EXPORT_TIMEOUT`                                                             | Export timeout for the batch processor.      |
| `OTEL_BLRP_MAX_QUEUE_SIZE`                                                             | Queue size limit.                            |
| `OTEL_LOGRECORD_ATTRIBUTE_COUNT_LIMIT`                                                 | Attribute count limit per log record.        |
| `OTEL_LOGRECORD_ATTRIBUTE_VALUE_LENGTH_LIMIT`                                          | Attribute value length limit per log record. |
