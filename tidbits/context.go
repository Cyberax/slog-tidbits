package tidbits

import (
	"context"
	"log/slog"
	"sync/atomic"
)

const loggerKey = "tidbitLogger"

var allowDefaultLoggerFallback = atomic.Bool{}

func EnableGlobalLoggerFallback(enabled bool) {
	allowDefaultLoggerFallback.Store(enabled)
}

func IsGlobalLoggerFallbackEnabled() bool {
	return allowDefaultLoggerFallback.Load()
}

func L(ctx context.Context) *slog.Logger {
	logger, ok := ctx.Value(loggerKey).(*slog.Logger)
	if ok {
		return logger
	}
	if !allowDefaultLoggerFallback.Load() {
		panic("no context logger and the global fallback is disabled")
	}
	return slog.Default()
}
