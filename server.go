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

var serversToShutdown = sync.WaitGroup{}

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

func waitServersShutdown() {
	slog.Info("waiting for servers to shutdown")
	serversToShutdown.Wait()
	slog.Info("all servers shutdown")
}
