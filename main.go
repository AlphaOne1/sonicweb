package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"github.com/corazawaf/coraza/v3"
	cor_http "github.com/corazawaf/coraza/v3/http"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/AlphaOne1/midgard"
	"github.com/AlphaOne1/midgard/handler"
)

func printLogo() {
	vcsRevision := "unknown"
	vcsTime := "unknown"
	vcsModified := ""
	goVersion := "unknown"

	if bi, biOK := debug.ReadBuildInfo(); biOK {
		goVersion = bi.GoVersion

		for _, s := range bi.Settings {
			switch s.Key {
			case "vcs.revision":
				vcsRevision = s.Value
			case "vcs.modified":
				if s.Value == "true" {
					vcsModified = "*"
				}
			case "vcs.time":
				vcsTime = s.Value
			}
		}
	}

	fmt.Println(`               /\______      `)
	fmt.Println(`            .-//\\     ''--__`)
	fmt.Println(`          /' // ||        _,/`)
	fmt.Println(`        /'  //__||      ,/   `)
	fmt.Println(`______/'    __         /____             _     _       __     __  `)
	fmt.Println(`\    /    /'_ '\      / ___/____  ____  (_)___| |     / /__  / /_ `)
	fmt.Println(` \  /    / / '\ \     \__ \/ __ \/ __ \/ / ___/ | /| / / _ \/ __ \`)
	fmt.Println(`  \/      / _  \     ___/ / /_/ / / / / / /__ | |/ |/ /  __/ /_/ /`)
	fmt.Println(`   \ .   / | \ /_   /____/\____/_/ /_/_/\___/ |__/|__/\___/_.___/ `)
	fmt.Println(`    \|\  |  \ // \       '\   `)
	fmt.Print(`     \ \-'__-/ _/ \        '\   `)
	fmt.Printf("Version: %v%v\n", vcsRevision, vcsModified)
	fmt.Print(`     @@_       _--/-----_    '\ `)
	fmt.Printf("     of: %v\n", vcsTime)
	fmt.Print(`        -------          ''-_  \`)
	fmt.Printf("  using: %v\n", goVersion)
	fmt.Println(`                             '-_\`)
	fmt.Println()
}

func setupLogging(logLevel string) {
	options := slog.HandlerOptions{
		AddSource: true,
		Level: func() slog.Level {
			var tmp slog.Level

			if levelErr := (&tmp).UnmarshalText([]byte(logLevel)); levelErr != nil {
				slog.Error("invalid loglevel",
					slog.String("error", levelErr.Error()),
					slog.String("given", logLevel))

				os.Exit(1)
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

	if ppid > 1 {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &options)))
	} else {
		slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &options)))
	}
}

func main() {
	printLogo()

	rootPath := flag.String("root", "/www", "root directory for webserver")
	basePath := flag.String("base", "/", "base path for serving")
	listenPort := flag.String("port", "8080", "port to listen on")
	listenAddress := flag.String("address", "", "address to listen on")
	instrumentPort := flag.Int("iport", 8081, "port to listen on for instrumentation")
	instrumentAddress := flag.String("iaddress", "", "address to listen on for instrumentation")
	enableTelemetry := flag.Bool("telemetry", true, "enable telemetry support")
	enablePprof := flag.Bool("pprof", false, "enable pprof support")
	logLevel := flag.String("log", "debug", "log level, valid options are debug, info, warn and error")

	flag.Parse()

	setupLogging(*logLevel)

	slog.Info("logging", slog.String("level", *logLevel))

	// termination handling
	termReceived := make(chan os.Signal, 1)
	signal.Notify(termReceived, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-termReceived
		slog.Info("received termination signal")
		os.Exit(0)
	}()

	// file path to serve from
	if rootPath != nil {
		slog.Info("using root directory", slog.String("root", *rootPath))
	} else {
		slog.Error("no root directory")
		os.Exit(1)
	}

	// base path in the URL to serve to
	if basePath != nil {
		slog.Info("using base path", slog.String("path", *basePath))
	} else {
		slog.Error("no basepath directory")
		os.Exit(1)
	}

	// setup opentelemetry and prometheus metricsExporter
	setupInstrumentation(instrumentAddress, instrumentPort, *enableTelemetry, *enablePprof)

	// First we initialize our waf and our seclang parser
	waf, wafErr := coraza.NewWAF(coraza.NewWAFConfig())

	// Now we parse our rules
	if wafErr != nil {
		slog.Error("could not initialize waf", slog.String("error", wafErr.Error()))
		os.Exit(1)
	}

	slog.Info("registering handler for FileServer")

	// remove all implicetely registered handlers
	http.DefaultServeMux = http.NewServeMux()

	mwStack := make([]midgard.Middleware, 0, 4)

	if *enableTelemetry {
		mwStack = append(mwStack, otelhttp.NewMiddleware("get_"+*basePath))
	}

	mwStack = append(mwStack,
		func(next http.Handler) http.Handler {
			return cor_http.WrapHandler(waf, next)
		},
		handler.Correlation,
		handler.AccessLogging,
		func(next http.Handler) http.Handler {
			return http.StripPrefix(*basePath, next)
		})

	http.Handle("GET "+*basePath,
		midgard.StackMiddlewareHandler(
			mwStack,
			http.FileServer(
				http.Dir(*rootPath),
			),
		))

	slog.Info("starting server")
	if err := http.ListenAndServe(*listenAddress+":"+*listenPort, nil); err != nil {
		slog.Error("error listening", slog.String("error", err.Error()))
		os.Exit(1)
	}

	os.Exit(0)
}
