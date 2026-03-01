// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Rasul Khiriev

// Package logger provides a thin wrapper around zerolog.Logger that adds
// convenience constructors and context-aware helpers used throughout the
// go-pass-keeper application.
//
// The Logger type embeds zerolog.Logger so all standard zerolog methods
// (Debug, Info, Warn, Error, Fatal, etc.) are available directly on *Logger.
// Application code should pass *Logger by pointer and obtain request-scoped
// loggers via FromContext or FromRequest.
package logger

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Logger is a thin wrapper around zerolog.Logger.
// Embedding zerolog.Logger exposes the full zerolog API while allowing the
// application to add helper methods without modifying the upstream type.
type Logger struct {
	zerolog.Logger
}

// NewLogger constructs a production-ready *Logger for the given role label
// (e.g. "server", "worker").
//
// The logger is configured with:
//   - global log level set to Debug (all levels are emitted);
//   - a "role" field set to role, useful for filtering logs from different
//     application components;
//   - a "ts" timestamp field added to every log entry;
//   - a "func" caller field that records the fully-qualified function name
//     (instead of the default file:line format) for easier log navigation.
//
// Output is written to os.Stdout in JSON format.
func NewLogger(role string) *Logger {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		return runtime.FuncForPC(pc).Name() // return function name
	}

	zerolog.CallerFieldName = "func"
	logger := zerolog.New(os.Stdout).With().
		Str("role", role).
		Timestamp().
		Caller().
		Logger()

	return &Logger{logger}
}
func NewClientLogger(role string) *Logger {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		return runtime.FuncForPC(pc).Name()
	}
	zerolog.CallerFieldName = "func"

	// Open log file near the executable
	execPath, _ := os.Executable()
	logPath := filepath.Join(filepath.Dir(execPath), "logs")
	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logFile = os.Stdout // fallback to stdout if file can't be opened
	}

	logger := zerolog.New(logFile).With().
		Str("role", role).
		Timestamp().
		Caller().
		Logger()

	return &Logger{logger}
}

// Nop returns a *Logger that discards all log output.
// It is intended for use in tests and other contexts where logging is
// undesirable or would produce noise.
func Nop() *Logger {
	return &Logger{zerolog.Nop()}
}

// GetChildLogger returns a new *Logger that inherits all fields of the
// receiver. The child logger can be enriched with additional context fields
// without affecting the parent logger.
func (l *Logger) GetChildLogger() *Logger {
	return &Logger{l.With().Logger()}
}

// FromRequest extracts the zerolog.Logger stored in the request's context by
// zerolog's log.Ctx helper and returns it as a *Logger.
//
// This is typically used in HTTP middleware that has previously attached a
// request-scoped logger to the context via zerolog's WithContext.
func FromRequest(r *http.Request) *Logger {
	return &Logger{*log.Ctx(r.Context())}
}

// FromContext extracts the zerolog.Logger stored in ctx by zerolog's log.Ctx
// helper and returns it as a *Logger.
//
// If no logger has been attached to ctx, zerolog returns its global logger,
// so this function never returns nil.
func FromContext(ctx context.Context) *Logger {
	return &Logger{*log.Ctx(ctx)}
}
