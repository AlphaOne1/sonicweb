// Copyright the SonicWeb contributors.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"context"
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	_ "time/tzdata"

	"go.uber.org/automaxprocs/maxprocs"

	"github.com/AlphaOne1/geany"
	"github.com/AlphaOne1/midgard"
	"github.com/AlphaOne1/midgard/defs"
	"github.com/AlphaOne1/midgard/handler/access_log"
	"github.com/AlphaOne1/midgard/handler/correlation"
	"github.com/AlphaOne1/midgard/util"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// ServerName is the reported server name in the header
const ServerName = "SonicWeb"

var buildInfoTag = ""  // buildInfoTag holds the tag information of the version control system
var exitFunc = os.Exit // exitFunc holds os.Exit for normal operations and is overridden for testing

//go:embed logo.tmpl
var logoTmpl string

// setupMaxProcs sets the maximum count of processors to use for scheduling
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

// generateFileHandler generates the handler to serve the files, initializing all necessary middlewares.
func generateFileHandler(
	enableTelemetry bool,
	enableTracing bool,
	basePath string,
	rootPath string,
	additionalHeaders [][2]string) http.Handler {

	mwStack := make([]defs.Middleware, 0, 4)

	if enableTelemetry || enableTracing {
		mwStack = append(mwStack, otelhttp.NewMiddleware("get_"+basePath))
	}

	mwStack = append(mwStack,
		wafMiddleware(nil),
		addHeaders(additionalHeaders),
		util.Must(correlation.New()),
		util.Must(access_log.New()),
		func(next http.Handler) http.Handler {
			return http.StripPrefix(basePath, next)
		})

	root, rootErr := os.OpenRoot(rootPath)

	if rootErr != nil {
		slog.Error("could not open root", slog.String("error", rootErr.Error()))
		exitFunc(1)
		return nil // silencing the static checker, unreachable
	}

	return midgard.StackMiddlewareHandler(
		mwStack,
		http.FileServerFS(
			root.FS(),
		),
	)
}

// MultiStringValue is used for command line parsing, holding values of repeated parameter occurrences
type MultiStringValue []string

// String returns the content as one string separated by comma, be careful, this is not a safe operation,
// if the parameters may contain comma themselves
func (m *MultiStringValue) String() string {
	return strings.Join(*m, ",")
}

// Set adds a new value.
func (m *MultiStringValue) Set(value string) error {
	*m = append(*m, value)
	return nil
}

// main initializes all necessary parts and starts the server.
func main() {
	_ = geany.PrintLogo(logoTmpl, map[string]string{"Tag": buildInfoTag})

	headersParam := &MultiStringValue{}
	headersFileParam := &MultiStringValue{}

	rootPath := flag.String("root", "/www", "root directory for webserver")
	basePath := flag.String("base", "/", "base path for serving")
	listenPort := flag.String("port", "8080", "port to listen on")
	listenAddress := flag.String("address", "", "address to listen on")
	flag.Var(headersParam, "header", "additional HTTP header")
	flag.Var(headersFileParam, "headerFile", "file containing additional HTTP headers")
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
		exitFunc(0)
	}

	setupLogging(*logLevel, *logStyle)
	setupMaxProcs()

	slog.Info("logging", slog.String("level", *logLevel))

	// termination handling
	termReceivedGlobal := make(chan os.Signal, 1)
	signal.Notify(termReceivedGlobal, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		s := <-termReceivedGlobal
		slog.Info("received termination signal")
		signalizeAll(s)
	}()

	slog.Info("using root directory", slog.String("root", *rootPath))

	if _, statErr := os.Stat(*rootPath); statErr != nil {
		slog.Error("could not get info of root path",
			slog.String("path", *rootPath),
			slog.String("error", statErr.Error()))
		exitFunc(1)
	}

	slog.Info("using base path", slog.String("path", *basePath))

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

	server := http.Server{
		Addr: *listenAddress + ":" + *listenPort,
	}

	defer func() { _ = server.Close() }()

	termReceived := make(chan os.Signal, 1)

	go func() {
		<-termReceived
		slog.Info("file server received termination signal")
		_ = server.Shutdown(context.Background())
	}()

	registerServer(FILE_SERVER, &server, &termReceived)

	handler := generateFileHandler(
		*enableTelemetry,
		len(*traceEndpoint) > 0,
		*basePath,
		*rootPath,
		append(headerParamToHeaders(*headersParam), headerFilesToHeaders(*headersFileParam)...))

	// remove all implicitly registered handlers
	http.DefaultServeMux = http.NewServeMux()
	http.Handle("GET "+*basePath, handler)

	slog.Info("starting server", slog.String("address", server.Addr))

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("error listening", slog.String("error", err.Error()))
		exitFunc(1)
	}

	slog.Info("server shutdown")
	exitFunc(0)
}
