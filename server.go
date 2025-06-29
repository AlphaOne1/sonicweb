package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// serverShutdownTimeout is the timeout given to the server to do a controlled shutdown.
const serverShutdownTimeout = 5 * time.Second

// serversToShutdown tracks the number of active servers being waited for to gracefully shut down.
var serversToShutdown = sync.WaitGroup{}

// waitServerShutdown gracefully shuts down the provided server upon receiving termination signals (SIGINT, SIGTERM).
// It logs the shutdown process, waits for an active server shutdown process to complete, and handles errors if any.
// Returns an error if the server fails to shut down gracefully.
func waitServerShutdown(server *http.Server, serverName string) error {
	serversToShutdown.Add(1)

	// termination handling
	termReceived := make(chan os.Signal, 1)
	signal.Notify(termReceived, syscall.SIGINT, syscall.SIGTERM)

	s := <-termReceived
	slog.Info("server received termination signal",
		slog.String("name", serverName),
		slog.String("signal", s.String()))

	shutdownCtx, shutdownCtxCancel := context.WithTimeout(context.Background(), serverShutdownTimeout)
	defer shutdownCtxCancel()

	shutdownErr := server.Shutdown(shutdownCtx)

	if shutdownErr != nil {
		slog.Error("error shutting down server",
			slog.String("name", serverName),
			slog.String("error", shutdownErr.Error()))
	} else {
		slog.Info("server shutdown", slog.String("name", serverName))
	}

	serversToShutdown.Done()

	if shutdownErr != nil {
		return fmt.Errorf("could not shutdown server %s: %w", serverName, shutdownErr)
	}

	return nil
}

// waitServersShutdown waits for all servers to complete their shutdown process using waitServerShutdown,
// before continuing execution.
func waitServersShutdown() {
	slog.Info("waiting for servers to shutdown")
	serversToShutdown.Wait()
	slog.Info("all servers shutdown")
}
