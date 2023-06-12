package tidbits

import (
	"context"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"testing"
)

func TestContextLogger(t *testing.T) {
	defer func() {
		EnableGlobalLoggerFallback(false)
		slog.SetDefault(slog.Default())
	}()

	EnableGlobalLoggerFallback(true)
	assert.True(t, IsGlobalLoggerFallbackEnabled())

	sink := NewSinkingLogger(slog.LevelInfo)
	slog.SetDefault(slog.New(sink.Handler()))
	L(context.Background()).Info("hello, world")
	assert.Equal(t, `{"time":"","level":"INFO","msg":"hello, world"}`, sink.Get())

	EnableGlobalLoggerFallback(false)
	assert.False(t, IsGlobalLoggerFallbackEnabled())

	assert.Panics(t, func() {
		L(context.Background()).Info("hello, world")
	}, "no context logger and the global fallback is disabled")
}
