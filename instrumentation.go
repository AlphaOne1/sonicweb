package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/http/pprof"
	"os"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
)

func setupInstrumentation(instrumentAddress *string, instrumentPort *int, enableTelemetry bool, enablePprof bool) {
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

func serveMetrics(address string, port int, enableTelemetry, enablePprof bool) {
	if !enableTelemetry && !enablePprof {
		return
	}

	listenAddress := fmt.Sprintf("%v:%v", address, port)

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

	if err := http.ListenAndServe(listenAddress, mux); err != nil {
		slog.Error("error serving metrics", slog.String("error", err.Error()))
		return
	}
}
