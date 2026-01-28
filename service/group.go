// SPDX-FileCopyrightText: 2026 The SonicWeb contributors.
// SPDX-License-Identifier: MPL-2.0

// Package service provides a Group type for managing multiple HTTP servers.
package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

// serverShutdownTimeout is the timeout given to the server to do a controlled shutdown.
const serverShutdownTimeout = 5 * time.Second

// ErrNoServers represents an error indicating that no servers have been configured.
var ErrNoServers = errors.New("no servers configured")

// ErrNilServer indicates that the server instance is nil and cannot be used.
var ErrNilServer = errors.New("server is nil")

// ErrServerNameLenMismatch indicates a mismatch between the lengths of the server and name values.
var ErrServerNameLenMismatch = errors.New("server and name length mismatch")

// Group represents a collection of HTTP servers managed together with shared lifecycle controls.
type Group struct {
	waitGroup       sync.WaitGroup
	procCount       atomic.Int32
	shutdownTimeout time.Duration
	log             *slog.Logger
	servers         []*http.Server
	serverNames     []string
}

// Option is a function that configures a Group by applying custom settings or modifications.
type Option func(*Group) error

// WithShutdownTimeout sets a timeout duration for server shutdown within the Group and returns an Option.
func WithShutdownTimeout(timeout time.Duration) Option {
	return func(g *Group) error {
		g.shutdownTimeout = timeout
		return nil
	}
}

// WithServer adds an HTTP server with a specified name to the group for management and lifecycle control.
func WithServer(server *http.Server, serverName string) Option {
	return func(g *Group) error {
		if server == nil {
			return fmt.Errorf("%w: %q", ErrNilServer, serverName)
		}

		g.servers = append(g.servers, server)
		g.serverNames = append(g.serverNames, serverName)

		return nil
	}
}

// WithServers adds the provided HTTP servers and their corresponding names to the Group configuration.
func WithServers(servers []*http.Server, serverNames []string) Option {
	return func(g *Group) error { //nolint:varnamelen // the name g is set in this file for groups, not changing here
		if len(servers) != len(serverNames) {
			return fmt.Errorf("%w: %d vs %d", ErrServerNameLenMismatch, len(servers), len(serverNames))
		}

		var errs []error

		for i, s := range servers {
			if s == nil {
				errs = append(errs, fmt.Errorf("%w: %q", ErrNilServer, serverNames[i]))
			}
		}

		if len(errs) > 0 {
			return errors.Join(errs...)
		}

		g.servers = append(g.servers, servers...)
		g.serverNames = append(g.serverNames, serverNames...)

		return nil
	}
}

// WithLogger sets a custom logger for the Group and returns an Option for configuration.
func WithLogger(log *slog.Logger) Option {
	return func(g *Group) error {
		g.log = log
		return nil
	}
}

// NewGroup creates and returns a new Group, applying the provided options. Returns an error if any option fails.
func NewGroup(options ...Option) (*Group, error) {
	group := &Group{
		shutdownTimeout: serverShutdownTimeout,
	}

	var errs []error

	for _, option := range options {
		if err := option(group); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return nil, fmt.Errorf("error creating group: %w", errors.Join(errs...))
	}

	if group.log == nil {
		group.log = slog.New(slog.DiscardHandler)
	}

	return group, nil
}

// ServerCount gives the current number of running servers in the group. This value is volatile and should be used
// only for informational purposes, e.g. display to the user.
func (g *Group) ServerCount() int {
	return int(g.procCount.Load())
}

// WaitAllServersShutdown waits for all running servers goroutines to complete and their shutdown processes using
// ShutdownWhenDone, before continuing execution.
func (g *Group) WaitAllServersShutdown() {
	g.log.Info("waiting for servers to shutdown")
	g.waitGroup.Wait()
	g.log.Info("all servers shutdown")
}

// StartAll starts all configured servers and returns once all listen sockets are bound.
// It does not block; use WaitAllServersShutdown() to wait for completion.
func (g *Group) StartAll(ctx context.Context) error {
	if ctx.Err() != nil {
		return fmt.Errorf("context error: %w", ctx.Err())
	}

	if ctx.Done() == nil {
		g.log.Debug("server group context is not cancellable, consider providing a cancellable context")
	}

	if len(g.servers) == 0 {
		return ErrNoServers
	}

	if len(g.servers) != len(g.serverNames) {
		return fmt.Errorf("%w: %d vs %d", ErrServerNameLenMismatch, len(g.servers), len(g.serverNames))
	}

	listeners, err := g.bindListeners(ctx)

	if err != nil {
		return err
	}

	// Start all servers after successful binding.
	for i := range len(g.servers) {
		listener := listeners[i]
		server := g.servers[i]
		serverName := g.serverNames[i]

		g.waitGroup.Add(1)
		g.procCount.Add(1)

		go g.handleServerCycle(ctx, server, listener, serverName)
	}

	return nil
}

// bindListeners binds network listeners for all configured servers and returns the listeners or an error if binding
// of any listener fails. On error, any listener bound already is closed.
func (g *Group) bindListeners(ctx context.Context) ([]net.Listener, error) {
	listeners := make([]net.Listener, 0, len(g.servers))

	// Bind all listeners first to guarantee "started".
	for serverIdx, server := range g.servers {
		addr := server.Addr

		if addr == "" {
			addr = ":http"
		}

		listener, err := (&net.ListenConfig{}).Listen(ctx, "tcp", addr)

		if err != nil {
			for _, l := range listeners {
				if closeErr := l.Close(); closeErr != nil {
					g.log.Error("error closing listener", slog.String("error", closeErr.Error()))
				}
			}

			return nil, fmt.Errorf("could not listen for server %s on %s: %w", g.serverNames[serverIdx], addr, err)
		}

		listeners = append(listeners, listener)
	}

	return listeners, nil
}

// handleServerCycle initializes and manages the lifecycle of a server, handling errors, shutdowns,
// and cancellations efficiently.
func (g *Group) handleServerCycle(ctx context.Context, server *http.Server, listener net.Listener, serverName string) {
	defer g.waitGroup.Done()
	defer g.procCount.Add(-1)

	serveErrCh := make(chan error, 1)

	go startServer(server, listener, serveErrCh)

	g.log.Info("server started",
		slog.String("name", serverName),
		slog.String("addr", listener.Addr().String()))

	select {
	case <-ctx.Done():
		g.log.Info("server received cancellation",
			slog.String("name", serverName),
			slog.String("reason", ctx.Err().Error()))

		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), g.shutdownTimeout)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			g.log.Error("error shutting down server",
				slog.String("name", serverName),
				slog.String("error", err.Error()))

			if closeErr := server.Close(); closeErr != nil {
				g.log.Error("error closing server", slog.String("error", closeErr.Error()))
			}
		} else {
			g.log.Info("server shut down", slog.String("name", serverName))
		}

		select {
		case <-time.After(serverShutdownTimeout):
			// this is a timeout applied _after_ the shutdown timeout of Shutdown
			g.log.Info("server shutdown timed out", slog.String("name", serverName))
		case err := <-serveErrCh:
			if err != nil {
				g.log.Error("server stopped with error",
					slog.String("name", serverName),
					slog.String("error", err.Error()))
			}
		}
	case err := <-serveErrCh:
		if err != nil {
			g.log.Error("server stopped with error",
				slog.String("name", serverName),
				slog.String("error", err.Error()))
		} else {
			g.log.Info("server stopped to accept new connections",
				slog.String("name", serverName))
		}
	}
}

// startServer handles the server lifecycle by starting it using the provided listener
// and sending any errors to the channel.
func startServer(server *http.Server, listener net.Listener, serveErrCh chan<- error) {
	if server.TLSConfig != nil {
		if err := server.ServeTLS(listener, "", ""); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			serveErrCh <- err
			return
		}
	} else {
		if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serveErrCh <- err
			return
		}
	}
	serveErrCh <- nil
}
