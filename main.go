// SPDX-FileCopyrightText: Copyright the SonicWeb contributors.
// SPDX-License-Identifier: MPL-2.0

// Package main contains the server logic of SonicWeb.
package main

import (
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
	_ "time/tzdata"

	"github.com/AlphaOne1/geany"
	"github.com/AlphaOne1/midgard"
	"github.com/AlphaOne1/midgard/defs"
	"github.com/AlphaOne1/midgard/handler/access_log"
	"github.com/AlphaOne1/midgard/handler/correlation"
	"github.com/AlphaOne1/midgard/util"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// ServerName is the reported server name in the header.
const ServerName = "SonicWeb"

// ReadTimeout is the timeout used to read header and body content.
const ReadTimeout = 2 * time.Second

var buildInfoTag = ""  // buildInfoTag holds the tag information of the version control system
var exitFunc = os.Exit // exitFunc holds os.Exit for normal operations and is overridden for testing

//go:embed logo.tmpl
var logoTmpl string

// MultiStringValue is used for command line parsing, holding values of repeated parameter occurrences.
type MultiStringValue []string

// String returns the content as one string separated by comma, be careful, this is not a safe operation
// if the parameters may contain comma themselves.
func (m *MultiStringValue) String() string {
	return strings.Join(*m, ",")
}

// Set adds a new value.
func (m *MultiStringValue) Set(value string) error {
	*m = append(*m, value)
	return nil
}

// ServerConfig holds all server configuration options.
type ServerConfig struct {
	RootPath          string
	BasePath          string
	ListenPort        string
	ListenAddress     string
	TLSCert           string
	TLSKey            string
	ClientCAs         *MultiStringValue
	AcmeDomains       *MultiStringValue
	CertCache         string
	AcmeEndpoint      string
	Headers           *MultiStringValue
	HeadersFiles      *MultiStringValue
	TryFiles          *MultiStringValue
	WafCfg            *MultiStringValue
	InstrumentPort    string
	InstrumentAddress string
	EnableTelemetry   bool
	TraceEndpoint     string
	EnablePprof       bool
	LogLevel          string
	LogStyle          string
	PrintVersion      bool
}

// setupFlags defines and parses all command line flags.
func setupFlags() ServerConfig {
	config := ServerConfig{
		ClientCAs:    &MultiStringValue{},
		AcmeDomains:  &MultiStringValue{},
		Headers:      &MultiStringValue{},
		HeadersFiles: &MultiStringValue{},
		TryFiles:     &MultiStringValue{},
		WafCfg:       &MultiStringValue{},
	}

	flag.StringVar(&config.RootPath, "root", "/www", "root directory for webserver")
	flag.StringVar(&config.BasePath, "base", "/", "base path for serving")
	flag.StringVar(&config.ListenPort, "port", "8080", "port to listen on")
	flag.StringVar(&config.ListenAddress, "address", "", "address to listen on")
	flag.StringVar(&config.TLSCert, "tlscert", "", "tls certificate file")
	flag.StringVar(&config.TLSKey, "tlskey", "", "tls key file")
	flag.Var(config.ClientCAs, "clientca", "client certificate authority file for mTLS")
	flag.Var(config.AcmeDomains, "acmedomain", "domain for automatic certificate retrieval")
	flag.StringVar(&config.CertCache, "certcache", os.TempDir(), "directory for certificate cache")
	flag.StringVar(&config.AcmeEndpoint, "acmeendpoint", "", " acme endpoint to use")
	flag.Var(config.Headers, "header", "additional HTTP header")
	flag.Var(config.HeadersFiles, "headerfile", "file containing additional HTTP headers")
	flag.Var(config.TryFiles, "tryfile", "always try to load file expression first")
	flag.Var(config.WafCfg, "wafcfg", "waf configuration file")
	flag.StringVar(&config.InstrumentPort, "iport", "8081", "port to listen on for instrumentation")
	flag.StringVar(&config.InstrumentAddress, "iaddress", "", "address to listen on for instrumentation")
	flag.BoolVar(&config.EnableTelemetry, "telemetry", true, "enable telemetry support")
	flag.StringVar(&config.TraceEndpoint, "trace-endpoint", "", "endpoint for tracing data")
	flag.BoolVar(&config.EnablePprof, "pprof", false, "enable pprof support")
	flag.StringVar(&config.LogLevel, "log", "info", "log level, valid options are debug, info, warn and error")
	flag.StringVar(&config.LogStyle, "logstyle", "auto", "log style, valid options are auto, text and json")
	flag.BoolVar(&config.PrintVersion, "version", false, "print version and exit")

	flag.Parse()

	return config
}

var errConversion = errors.New("conversion error")

// generateFileHandler generates the handler to serve the files, initializing all necessary middlewares.
func generateFileHandler(
	enableTelemetry bool,
	enableTracing bool,
	basePath string,
	rootPath string,
	additionalHeaders [][2]string,
	tryFiles []string,
	wafCfg []string) (http.Handler, error) {

	mwStack := make([]defs.Middleware, 0, 4)

	if enableTelemetry || enableTracing {
		mwStack = append(mwStack, otelhttp.NewMiddleware("get_"+basePath))
	}

	root, rootErr := os.OpenRoot(rootPath)

	if rootErr != nil {
		return nil, fmt.Errorf("could not open root: %w", rootErr) // silencing the static checker, unreachable
	}

	if len(wafCfg) > 0 {
		wafMW, wafMWErr := wafMiddleware(wafCfg)

		if wafMWErr != nil {
			return nil, fmt.Errorf("could not initialize waf middleware: %w", wafMWErr)
		}

		mwStack = append(mwStack, wafMW)
	}

	statFS, statFSOK := root.FS().(fs.StatFS)

	if !statFSOK {
		return nil, fmt.Errorf("could not get StatFS from RootFS: %w", errConversion)
	}

	mwStack = append(mwStack,
		addHeaders(additionalHeaders),
		util.Must(correlation.New()),
		util.Must(access_log.New()),
		addTryFiles(tryFiles, statFS),
		checkValidFilePath(),
		func(next http.Handler) http.Handler {
			return http.StripPrefix(basePath, next)
		})

	return midgard.StackMiddlewareHandler(
		mwStack,
		http.FileServerFS(
			root.FS(),
		),
	), nil
}

// main initializes all necessary parts and starts the server.
func main() {
	startInit := time.Now()
	_ = geany.PrintLogo(logoTmpl, map[string]string{"Tag": buildInfoTag})

	// Parse command line flags
	config := setupFlags()

	if config.PrintVersion {
		// we already printed the logo that contains all the necessary information
		exitFunc(0)
	}

	if err := setupLogging(config.LogLevel, config.LogStyle); err != nil {
		slog.Error("error setting up logging", slog.String("error", err.Error()))
		exitFunc(1)
	}

	slog.Info("logging", slog.String("level", config.LogLevel))

	slog.Info("using root directory", slog.String("root", config.RootPath))

	if _, statErr := os.Stat(config.RootPath); statErr != nil {
		slog.Error("could not get info of root path",
			slog.String("path", config.RootPath),
			slog.String("error", statErr.Error()))
		exitFunc(1)
	}

	slog.Info("using base path", slog.String("path", config.BasePath))

	if len(config.TraceEndpoint) > 0 {
		if _, err := initTracer(config.TraceEndpoint); err != nil {
			slog.Error("could not initialize tracing", slog.String("error", err.Error()))
			exitFunc(1)
		}

		slog.Info("tracing initialized")
	} else {
		slog.Info("tracing disabled")
	}

	slog.Info("registering handler for FileServer")

	tlsConfig, tlsConfigErr := generateTLSConfig(
		config.TLSCert,
		config.TLSKey,
		*config.AcmeDomains,
		config.CertCache,
		config.AcmeEndpoint,
		*config.ClientCAs)

	if tlsConfigErr != nil {
		slog.Error("invalid TLS configuration", slog.String("error", tlsConfigErr.Error()))
		exitFunc(1)
	}

	server := http.Server{
		Addr:              net.JoinHostPort(config.ListenAddress, config.ListenPort),
		ReadHeaderTimeout: ReadTimeout,
		ReadTimeout:       ReadTimeout,
		TLSConfig:         tlsConfig,
	}

	defer func() { _ = server.Close() }()

	headers, headersErr := headerFilesToHeaders(*config.HeadersFiles)

	if headersErr != nil {
		slog.Error("could not process headers file",
			slog.Any("files", *config.HeadersFiles),
			slog.String("error", headersErr.Error()))
		exitFunc(1)
	}

	handler, handlerErr := generateFileHandler(
		config.EnableTelemetry,
		len(config.TraceEndpoint) > 0,
		config.BasePath,
		config.RootPath,
		append(headerParamToHeaders(*config.Headers), headers...),
		*config.TryFiles,
		*config.WafCfg)

	if handlerErr != nil {
		slog.Error("could not generate file handler", slog.String("error", handlerErr.Error()))
		exitFunc(1)
	}

	// remove all implicitly registered handlers
	http.DefaultServeMux = http.NewServeMux()
	http.Handle("GET "+config.BasePath, handler)

	go func() {
		var listenErr error

		if tlsConfig != nil {
			slog.Info("starting tls server",
				slog.String("address", server.Addr),
				slog.Duration("t_init", time.Since(startInit)),
				slog.String("cert", config.TLSCert),
				slog.String("key", config.TLSKey),
				slog.Any("acmeDomains", *config.AcmeDomains),
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
	}()

	// set up opentelemetry with prometheus metricsExporter
	setupMetricsInstrumentation(
		&config.InstrumentAddress,
		&config.InstrumentPort,
		config.EnableTelemetry,
		config.EnablePprof)

	fileServerShutdownErr := waitServerShutdown(&server, "file")

	if fileServerShutdownErr != nil {
		slog.Error("error shutting down server", slog.String("error", fileServerShutdownErr.Error()))
	}

	waitServersShutdown()

	exitFunc(0)
}
