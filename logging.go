// SPDX-FileCopyrightText: 2026 The SonicWeb contributors.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"slices"
)

var errLogConfig = errors.New("invalid log configuration")

// setupLogging sets the log format and level. It can try to guess in which environment
// SonicWeb runs (logStyle "auto"). If its parent seems to not be an init process, then
// text logging is used, otherwise JSON.
func setupLogging(logLevel string, logStyle string) error {
	var parsedLogLevel slog.Level

	if levelErr := (&parsedLogLevel).UnmarshalText([]byte(logLevel)); levelErr != nil {
		return fmt.Errorf("invalid loglevel: %w", levelErr)
	}

	options := slog.HandlerOptions{
		AddSource: func() bool { return parsedLogLevel <= slog.LevelDebug }(),
		Level:     parsedLogLevel,
		ReplaceAttr: func(_ /*groups*/ []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				a.Value = slog.StringValue(a.Value.Time().Format("2006-01-02T15:04:05.000000Z07:00"))
			}

			return a
		},
	}

	ppid := os.Getppid()

	switch {
	case (logStyle == "auto" && ppid > 1) || logStyle == "text":
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &options)))
	case logStyle == "auto" || logStyle == "json":
		options.ReplaceAttr = nil
		slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &options)))
	default:
		return fmt.Errorf("unsupported log style %s: %w", logStyle, errLogConfig)
	}

	return nil
}

type MultiHandler struct {
	handlers []slog.Handler
}

func NewMultiHandler(handlers ...slog.Handler) *MultiHandler {
	return &MultiHandler{handlers: slices.DeleteFunc(
		handlers,
		func(h slog.Handler) bool {
			return h == nil
		})}
}

func (t *MultiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, h := range t.handlers {
		if h.Enabled(ctx, level) {
			return true
		}
	}

	return false
}

func (t *MultiHandler) Handle(ctx context.Context, r slog.Record) error {
	var errs []error

	for _, h := range t.handlers {
		c := r.Clone()
		if err := h.Handle(ctx, c); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func (t *MultiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, 0, len(t.handlers))

	for _, h := range t.handlers {
		handlers = append(handlers, h.WithAttrs(attrs))
	}

	return &MultiHandler{handlers: handlers}
}

func (t *MultiHandler) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, 0, len(t.handlers))

	for _, h := range t.handlers {
		handlers = append(handlers, h.WithGroup(name))
	}

	return &MultiHandler{handlers: handlers}
}
