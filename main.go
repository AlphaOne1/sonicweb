// Copyright the SonicWeb contributors.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"context"
	_ "embed"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	_ "time/tzdata"

	"go.uber.org/automaxprocs/maxprocs"

	"github.com/AlphaOne1/midgard"
	"github.com/AlphaOne1/midgard/defs"
	"github.com/AlphaOne1/midgard/handler/access_log"
	"github.com/AlphaOne1/midgard/handler/correlation"
	"github.com/AlphaOne1/midgard/util"

	"github.com/corazawaf/coraza/v3"
	corhttp "github.com/corazawaf/coraza/v3/http"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

const ServerName = "SonicWeb"

var buildInfoTag = ""
var exitFunc = os.Exit

//go:embed logo.tmpl
var logoTmpl string

func setupLogging(logLevel string, logStyle string) {
	options := slog.HandlerOptions{
		AddSource: true,
		Level: func() slog.Level {
			var tmp slog.Level

			if levelErr := (&tmp).UnmarshalText([]byte(logLevel)); levelErr != nil {
				slog.Error("invalid loglevel",
					slog.String("error", levelErr.Error()),
					slog.String("given", logLevel))

				exitFunc(1)
			}

			return tmp
		}(),
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				a.Value = slog.StringValue(a.Value.Time().Format("2006-01-02T15:04:05.000000"))
			}
			return a
		},
	}

	ppid := os.Getppid()

	if (logStyle == "auto" && ppid > 1) || logStyle == "text" {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &options)))
	} else if logStyle == "auto" || logStyle == "json" {
		slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &options)))
	} else {
		slog.Error("unsupported log style", slog.String("logStyle", logStyle))
		exitFunc(1)
	}
}

func setupMaxProcs() {
	if _, mpFound := os.LookupEnv("GOMAXPROCS"); !mpFound {
		if _, err := maxprocs.Set(maxprocs.Logger(func(format string, args ...any) {
			if slog.Default().Enabled(context.Background(), slog.LevelInfo) {
				message := fmt.Sprintf(format, args...)
				slog.Info(message)
			}
		})); err != nil {
			slog.Error("failed to automatically set GOMAXPROCS",
				slog.String("error", err.Error()))
			exitFunc(1)
		}
	}
}

func generateFileHandler(
	enableTelemetry bool,
	enableTracing bool,
	basePath string,
	rootPath string) http.Handler {

	// First we initialize our waf and our seclang parser
	waf, wafErr := coraza.NewWAF(coraza.NewWAFConfig())

	// Now we parse our rules
	if wafErr != nil {
		slog.Error("could not initialize waf", slog.String("error", wafErr.Error()))
		exitFunc(1)
	}

	mwStack := make([]defs.Middleware, 0, 4)

	if enableTelemetry || enableTracing {
		mwStack = append(mwStack, otelhttp.NewMiddleware("get_"+basePath))
	}

	mwStack = append(mwStack,
		func(next http.Handler) http.Handler {
			return corhttp.WrapHandler(waf, next)
		},
		func(next http.Handler) http.Handler {
			serverVal := ServerName

			if len(buildInfoTag) > 0 {
				serverVal = serverVal + "/" + buildInfoTag
			}

			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Server", serverVal)
				next.ServeHTTP(w, r)
			})
		},
		util.Must(correlation.New()),
		util.Must(access_log.New()),
		func(next http.Handler) http.Handler {
			return http.StripPrefix(basePath, next)
		})

	return midgard.StackMiddlewareHandler(
		mwStack,
		http.FileServerFS(
			os.DirFS(rootPath),
		),
	)
}

func main() {
	PrintLogo(logoTmpl, map[string]string{"Tag": buildInfoTag})

	rootPath := flag.String("root", "/www", "root directory for webserver")
	basePath := flag.String("base", "/", "base path for serving")
	listenPort := flag.String("port", "8080", "port to listen on")
	listenAddress := flag.String("address", "", "address to listen on")
	instrumentPort := flag.Int("iport", 8081, "port to listen on for instrumentation")
	instrumentAddress := flag.String("iaddress", "", "address to listen on for instrumentation")
	enableTelemetry := flag.Bool("telemetry", true, "enable telemetry support")
	traceEndpoint := flag.String("trace-endpoint", "", "endpoint for tracing data")
	enablePprof := flag.Bool("pprof", false, "enable pprof support")
	logLevel := flag.String("log", "info", "log level, valid options are debug, info, warn and error")
	logStyle := flag.String("logstyle", "auto", "log style, valid options are auto, text and json")
	printVersion := flag.Bool("version", false, "print version and exit")

	flag.Parse()

	if *printVersion {
		// we already printed the logo, that contains all the necessary information
		os.Exit(0)
	}

	setupLogging(*logLevel, *logStyle)
	setupMaxProcs()

	slog.Info("logging", slog.String("level", *logLevel))

	// termination handling
	termReceived := make(chan os.Signal, 1)
	signal.Notify(termReceived, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-termReceived
		slog.Info("received termination signal")
		exitFunc(0)
	}()

	// file path to serve from
	if rootPath != nil {
		slog.Info("using root directory", slog.String("root", *rootPath))
	} else {
		slog.Error("no root directory")
		exitFunc(1)
	}

	if _, statErr := os.Stat(*rootPath); statErr != nil {
		slog.Error("could not get info of root path",
			slog.String("path", *rootPath),
			slog.String("error", statErr.Error()))
		exitFunc(1)
	}

	// base path in the URL to serve to
	if basePath != nil {
		slog.Info("using base path", slog.String("path", *basePath))
	} else {
		slog.Error("no basepath directory")
		exitFunc(1)
	}

	if len(*traceEndpoint) > 0 {
		if _, err := initTracer(*traceEndpoint); err != nil {
			slog.Error("could not initialize tracing", slog.String("error", err.Error()))
			exitFunc(1)
		}

		slog.Info("tracing initialized")
	} else {
		slog.Info("tracing disabled")
	}

	// setup opentelemetry with prometheus metricsExporter
	setupMetricsInstrumentation(instrumentAddress, instrumentPort, *enableTelemetry, *enablePprof)

	slog.Info("registering handler for FileServer")

	// remove all implicitly registered handlers
	http.DefaultServeMux = http.NewServeMux()

	handler := generateFileHandler(*enableTelemetry, len(*traceEndpoint) > 0, *basePath, *rootPath)

	http.Handle("GET "+*basePath, handler)

	slog.Info("starting server")
	if err := http.ListenAndServe(*listenAddress+":"+*listenPort, nil); err != nil {
		slog.Error("error listening", slog.String("error", err.Error()))
		exitFunc(1)
	}

	exitFunc(0)
}
