package lhelper

import (
	"context"
	"github.com/Cyberax/slog-tidbits/tidbits"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"testing"
)

func TestContextLogger(t *testing.T) {
	defer func() {
		EnableGlobalLoggerFallback(false)
		slog.SetDefault(slog.Default())
	}()

	assert.Nil(t, TryGetLoggerFromContext(context.Background()))

	EnableGlobalLoggerFallback(true)
	assert.True(t, IsGlobalLoggerFallbackEnabled())

	sink := tidbits.NewSinkingLogger(slog.LevelInfo)
	slog.SetDefault(slog.New(sink.Handler()))
	L(context.Background()).Info("hello, world")
	assert.Equal(t, `{"time":"","level":"INFO","msg":"hello, world"}`, sink.Get())

	EnableGlobalLoggerFallback(false)
	assert.False(t, IsGlobalLoggerFallbackEnabled())

	assert.Panics(t, func() {
		L(context.Background()).Info("hello, world")
	}, "no context logger and the global fallback is disabled")

	ctx := WithLogger(context.Background(), slog.New(sink.Handler()))
	L(ctx).Info("hello, world")
	assert.Equal(t, `{"time":"","level":"INFO","msg":"hello, world"}`, sink.Get())
	assert.NotNil(t, TryGetLoggerFromContext(ctx))
}
