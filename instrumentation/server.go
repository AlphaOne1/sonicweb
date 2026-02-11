// SPDX-FileCopyrightText: 2026 The SonicWeb contributors.
// SPDX-License-Identifier: MPL-2.0

package instrumentation

import (
	"log/slog"
	"net"
	"net/http"
	"net/http/pprof"
	"time"
)

// ReadTimeout defines the maximum duration for reading the entire request, including the body, from the client.
const ReadTimeout = 2 * time.Second

// WriteTimeout defines the maximum duration before timing out writes of the response. Adjusted to support the standard
// 30 seconds CPU profiles of pprof.
const WriteTimeout = 35 * time.Second

// IdleTimeout defines the maximum amount of time to wait for the next request when keep-alives are enabled.
const IdleTimeout = 30 * time.Second

// MaxHeaderBytes limits the size of request headers to mitigate memory abuse.
const MaxHeaderBytes = 1 << 20 // 1 MiB

// Server sets up the metrics functionality. It opens a separate port, and metrics collectors can fetch their
// data from there. For profiling purposes, if enabled, also the pprof endpoints are added on the same port.
func Server(
	address string,
	port string,
	metricHandler http.Handler,
	enablePprof bool,
	log *slog.Logger) (*http.Server, error) {

	if log == nil {
		// get a "do nothing" logger if none is set
		log = slog.New(slog.DiscardHandler)
	}

	if metricHandler == nil && !enablePprof {
		return nil, nil //nolint:nilnil // this can totally happen
	}

	listenAddress := net.JoinHostPort(address, port)

	mux := http.NewServeMux()

	if enablePprof {
		host := net.ParseIP(address)
		isLoopback := address == "localhost" || (host != nil && host.IsLoopback())

		if !isLoopback {
			log.Warn("pprof requested but listen address is not loopback, ensure this port is not publically exposed",
				slog.String("address", listenAddress))
		}

		log.Info("serving pprof", slog.String("address", listenAddress+"/debug/pprof"))
		mux.HandleFunc("GET /debug/pprof/", pprof.Index)
		mux.HandleFunc("GET /debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("GET /debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("GET /debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("GET /debug/pprof/trace", pprof.Trace)

		// Common profiles expected under /debug/pprof/<name>
		for _, i := range []string{"allocs", "block", "goroutine", "heap", "mutex", "threadcreate"} {
			mux.Handle("GET /debug/pprof/"+i, pprof.Handler(i))
		}
	} else {
		log.Info("serving pprof disabled")
	}

	if metricHandler != nil {
		log.Info("serving metrics", slog.String("address", listenAddress+"/metrics"))
		mux.Handle("GET /metrics", metricHandler)
	} else {
		log.Info("serving metrics disabled")
	}

	server := http.Server{
		Addr:              listenAddress,
		Handler:           mux,
		MaxHeaderBytes:    MaxHeaderBytes,
		IdleTimeout:       IdleTimeout,
		ReadHeaderTimeout: ReadTimeout,
		ReadTimeout:       ReadTimeout,
		WriteTimeout:      WriteTimeout,
	}

	return &server, nil
}
