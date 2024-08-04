package lhelper

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"sync/atomic"
	"time"
)

const LevelTrace = slog.Level(-10)

const loggerKey = "tidbitLogger"

var allowDefaultLoggerFallback = atomic.Bool{}

func EnableGlobalLoggerFallback(enabled bool) {
	allowDefaultLoggerFallback.Store(enabled)
}

func IsGlobalLoggerFallbackEnabled() bool {
	return allowDefaultLoggerFallback.Load()
}

func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

func L(ctx context.Context) *slog.Logger {
	logger, ok := ctx.Value(loggerKey).(*slog.Logger)
	if ok {
		return slog.New(contextualizedLog{
			suppliedCtx: ctx,
			delegate:    logger.Handler(),
		})
	}
	if !allowDefaultLoggerFallback.Load() {
		panic("no context logger and the global fallback is disabled")
	}
	return slog.New(contextualizedLog{
		suppliedCtx: ctx,
		delegate:    slog.Default().Handler(),
	})
}

func TryGetLoggerFromContext(ctx context.Context) *slog.Logger {
	logger, ok := ctx.Value(loggerKey).(*slog.Logger)
	if ok {
		return logger
	}
	return nil
}

// LTRACE helper logs a trace message. It's a shorthand for L(ctx).Log(LevelTrace, msg, args...).
// It's not a wrapper function, but a reimplementation of slog.Logger.log to make sure it gets the
// correct PC location for the caller.
func LTRACE(ctx context.Context, msg string, args ...any) {
	logger := L(ctx)
	if !logger.Enabled(ctx, LevelTrace) {
		return
	}
	var pc uintptr
	var pcs [1]uintptr
	// skip [runtime.Callers, this function]
	runtime.Callers(2, pcs[:])
	pc = pcs[0]

	r := slog.NewRecord(time.Now(), LevelTrace, msg, pc)
	r.Add(args...)
	_ = logger.Handler().Handle(ctx, r)
}

type contextualizedLog struct {
	suppliedCtx context.Context
	delegate    slog.Handler
}

var _ slog.Handler = &contextualizedLog{}

func (c contextualizedLog) ensureContext(ctx context.Context) context.Context {
	stringer, ok := ctx.(fmt.Stringer)
	if ok && stringer.String() == "context.Background" {
		return c.suppliedCtx
	}
	return ctx
}

func (c contextualizedLog) Enabled(ctx context.Context, level slog.Level) bool {
	return c.delegate.Enabled(c.ensureContext(ctx), level)
}

func (c contextualizedLog) Handle(ctx context.Context, record slog.Record) error {
	return c.delegate.Handle(c.ensureContext(ctx), record)
}

func (c contextualizedLog) WithAttrs(attrs []slog.Attr) slog.Handler {
	return contextualizedLog{
		suppliedCtx: c.suppliedCtx,
		delegate:    c.delegate.WithAttrs(attrs),
	}
}

func (c contextualizedLog) WithGroup(name string) slog.Handler {
	return contextualizedLog{
		suppliedCtx: c.suppliedCtx,
		delegate:    c.delegate.WithGroup(name),
	}
}
