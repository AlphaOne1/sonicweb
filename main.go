// Copyright the SonicWeb contributors.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"context"
	"crypto/tls"
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	_ "time/tzdata"

	"github.com/AlphaOne1/geany"
	"github.com/AlphaOne1/midgard"
	"github.com/AlphaOne1/midgard/defs"
	"github.com/AlphaOne1/midgard/handler/access_log"
	"github.com/AlphaOne1/midgard/handler/correlation"
	"github.com/AlphaOne1/midgard/util"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/automaxprocs/maxprocs"
	"golang.org/x/crypto/acme/autocert"
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

func generateTLSConfig(cert string, key string, acmeDomains []string, certCache string) (*tls.Config, error) {
	if (len(cert) > 0) != (len(key) > 0) {
		return nil, fmt.Errorf("invalid tls config, cert and key must both be given or not given")
	}

	if len(cert) > 0 && len(acmeDomains) > 0 {
		return nil, fmt.Errorf("either cert+key or acmeDomains are to be given")
	}

	if len(cert) > 0 {
		cert, err := tls.LoadX509KeyPair(cert, key)

		if err != nil {
			return nil, fmt.Errorf("could not load certificate: %w\n", err)
		}

		return &tls.Config{
			Certificates: []tls.Certificate{cert},
		}, nil
	}

	if len(acmeDomains) > 0 {
		// automatic certificate management with autocert
		certManager := autocert.Manager{
			Cache:      autocert.DirCache(certCache),
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(acmeDomains...),
		}

		return certManager.TLSConfig(), nil
	}

	// completely valid, we do not have a TLS config
	return nil, nil
}

// generateFileHandler generates the handler to serve the files, initializing all necessary middlewares.
func generateFileHandler(
	enableTelemetry bool,
	enableTracing bool,
	basePath string,
	rootPath string,
	additionalHeaders [][2]string,
	tryFiles []string,
	wafCfg []string) http.Handler {

	mwStack := make([]defs.Middleware, 0, 4)

	if enableTelemetry || enableTracing {
		mwStack = append(mwStack, otelhttp.NewMiddleware("get_"+basePath))
	}

	root, rootErr := os.OpenRoot(rootPath)

	if rootErr != nil {
		slog.Error("could not open root", slog.String("error", rootErr.Error()))
		exitFunc(1)
		return nil // silencing the static checker, unreachable
	}

	mwStack = append(mwStack,
		wafMiddleware(wafCfg),
		addHeaders(additionalHeaders),
		util.Must(correlation.New()),
		util.Must(access_log.New()),
		addTryFiles(tryFiles, root.FS().(fs.StatFS)),
		func(next http.Handler) http.Handler {
			return http.StripPrefix(basePath, next)
		})

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
	startInit := time.Now()
	_ = geany.PrintLogo(logoTmpl, map[string]string{"Tag": buildInfoTag})

	headersParam := &MultiStringValue{}
	headersFileParam := &MultiStringValue{}
	tryFiles := &MultiStringValue{}
	wafCfg := &MultiStringValue{}
	acmeDomains := &MultiStringValue{}

	rootPath := flag.String("root", "/www", "root directory for webserver")
	basePath := flag.String("base", "/", "base path for serving")
	listenPort := flag.String("port", "8080", "port to listen on")
	listenAddress := flag.String("address", "", "address to listen on")
	tlsCert := flag.String("tlscert", "", "tls certificate file")
	tlsKey := flag.String("tlskey", "", "tls key file")
	flag.Var(acmeDomains, "acmedomain", "domain for automatic certificate retrieval")
	certCache := flag.String("certcache", os.TempDir(), "directory for certificate cache")
	flag.Var(headersParam, "header", "additional HTTP header")
	flag.Var(headersFileParam, "headerfile", "file containing additional HTTP headers")
	flag.Var(tryFiles, "tryfile", "always try to load file expression first")
	flag.Var(wafCfg, "wafcfg", "waf configuration file")
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

	tlsConfig, tlsConfigErr := generateTLSConfig(*tlsCert, *tlsKey, *acmeDomains, *certCache)

	if tlsConfigErr != nil {
		slog.Error("invalid TLS configuration", slog.String("error", tlsConfigErr.Error()))
		exitFunc(1)
	}

	server := http.Server{
		Addr:      *listenAddress + ":" + *listenPort,
		TLSConfig: tlsConfig,
	}

	defer func() { _ = server.Close() }()

	termReceived := make(chan os.Signal, 1)

	go func() {
		<-termReceived
		slog.Info("file server received termination signal")
		_ = server.Shutdown(context.Background())
	}()

	// termination handling
	signal.Notify(termReceived, syscall.SIGINT, syscall.SIGTERM)

	handler := generateFileHandler(
		*enableTelemetry,
		len(*traceEndpoint) > 0,
		*basePath,
		*rootPath,
		append(headerParamToHeaders(*headersParam), headerFilesToHeaders(*headersFileParam)...),
		*tryFiles,
		*wafCfg)

	// remove all implicitly registered handlers
	http.DefaultServeMux = http.NewServeMux()
	http.Handle("GET "+*basePath, handler)

	var listenErr error

	if tlsConfig != nil {
		slog.Info("starting tls server",
			slog.String("address", server.Addr),
			slog.Duration("t_init", time.Since(startInit)),
			slog.String("cert", *tlsCert),
			slog.String("key", *tlsKey),
			slog.Any("acmeDomains", *acmeDomains),
		)

		listenErr = server.ListenAndServeTLS("", "")
	} else {
		slog.Info("starting server",
			slog.String("address", server.Addr),
			slog.Duration("t_init", time.Since(startInit)))

		listenErr = server.ListenAndServe()
	}

	if listenErr != nil && !errors.Is(listenErr, http.ErrServerClosed) {
		slog.Error("error listening", slog.String("error", listenErr.Error()))
		exitFunc(1)
	}

	slog.Info("server shutdown")
	exitFunc(0)
}
