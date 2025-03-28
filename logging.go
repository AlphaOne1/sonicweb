// Copyright the SonicWeb contributors.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"log/slog"
	"os"
)

// setupLogging sets the log format and level. It can try to guess in which environment
// SonicWeb runs (logStyle "auto"). If its parent seems to not be an init process, then
// text logging is used, otherwise JSON.
func setupLogging(logLevel string, logStyle string) {
	var parsedLogLevel slog.Level

	if levelErr := (&parsedLogLevel).UnmarshalText([]byte(logLevel)); levelErr != nil {
		slog.Error("invalid loglevel",
			slog.String("error", levelErr.Error()),
			slog.String("given", logLevel))

		exitFunc(1)
	}

	options := slog.HandlerOptions{
		AddSource: func() bool { return parsedLogLevel <= slog.LevelDebug }(),
		Level:     parsedLogLevel,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				a.Value = slog.StringValue(a.Value.Time().Format("2006-01-02T15:04:05.000000Z07:00"))
			}
			return a
		},
	}

	ppid := os.Getppid()

	if (logStyle == "auto" && ppid > 1) || logStyle == "text" {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &options)))
	} else if logStyle == "auto" || logStyle == "json" {
		options.ReplaceAttr = nil
		slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &options)))
	} else {
		slog.Error("unsupported log style", slog.String("logStyle", logStyle))
		exitFunc(1)
	}
}
