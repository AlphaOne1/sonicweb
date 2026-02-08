// SPDX-FileCopyrightText: 2026 The SonicWeb contributors.
// SPDX-License-Identifier: MPL-2.0

package instrumentation

import (
	"context"
	"errors"
	"log/slog"
	"slices"
)

// MultiHandler is a composite handler that forwards log records to multiple underlying handlers.
type MultiHandler struct {
	handlers []slog.Handler
}

// NewMultiHandler creates a MultiHandler that delegates log records to multiple provided slog.Handler instances.
func NewMultiHandler(handlers ...slog.Handler) *MultiHandler {
	tmp := slices.Clone(handlers)
	tmp = slices.DeleteFunc(
		tmp,
		func(h slog.Handler) bool {
			return h == nil
		})

	return &MultiHandler{handlers: tmp}
}

// Enabled determines if any underlying handler is enabled for the given context and log level.
func (t *MultiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, h := range t.handlers {
		if h.Enabled(ctx, level) {
			return true
		}
	}

	return false
}

// Handle processes a log record by forwarding it to all underlying handlers and aggregates any errors encountered.
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

// WithAttrs returns a new MultiHandler with the specified attributes added to all underlying handlers.
func (t *MultiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, 0, len(t.handlers))

	for _, h := range t.handlers {
		handlers = append(handlers, h.WithAttrs(attrs))
	}

	return &MultiHandler{handlers: handlers}
}

// WithGroup returns a new MultiHandler with the specified group name applied to all underlying handlers.
func (t *MultiHandler) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, 0, len(t.handlers))

	for _, h := range t.handlers {
		handlers = append(handlers, h.WithGroup(name))
	}

	return &MultiHandler{handlers: handlers}
}
