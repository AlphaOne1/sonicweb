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
	slog.Info(fmt.Sprintf("%s server received termination signal", serverName),
		slog.String("signal", s.String()))

	shutdownCtx, shutdownCtxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCtxCancel()

	shutdownErr := server.Shutdown(shutdownCtx)

	if shutdownErr != nil {
		slog.Error(fmt.Sprintf("error shutting down %s server", serverName),
			slog.String("error", shutdownErr.Error()))
	} else {
		slog.Info(fmt.Sprintf("%s server shutdown", serverName))
	}

	serversToShutdown.Done()
	return shutdownErr
}

// waitServersShutdown waits for all servers to complete their shutdown process using waitServerShutdown,
// before continuing execution.
func waitServersShutdown() {
	slog.Info("waiting for servers to shutdown")
	serversToShutdown.Wait()
	slog.Info("all servers shutdown")
}
