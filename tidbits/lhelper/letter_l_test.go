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

func TestLTRACE(t *testing.T) {
	sink := tidbits.NewSinkingLogger(LevelTrace)
	slog.SetDefault(slog.New(sink.Handler()))
	ctx := WithLogger(context.Background(), slog.New(sink.Handler()))
	LTRACE(ctx, "hello, world", "this", "is a test")
	assert.Equal(t, `{"time":"","level":"DEBUG-6","msg":"hello, world","this":"is a test"}`, sink.Get())
}

type TestExtractor struct {
}

var _ tidbits.ContextExtractor = &TestExtractor{}

func (t *TestExtractor) MergeContextAttrs(ctx context.Context, curAttrs []slog.Attr) []slog.Attr {
	val := ctx.Value("TestValue").(string)
	return append(curAttrs, slog.String("TestValue", val))
}

func TestExtractors(t *testing.T) {
	sink := tidbits.NewSinkingLogger(slog.LevelInfo)
	conv := tidbits.NewSlogConvenience(tidbits.SlogOptions{
		Extractors: []tidbits.ContextExtractor{&TestExtractor{}},
	}, sink.Handler())

	ctx := WithLogger(context.Background(), slog.New(conv))
	ctx = context.WithValue(ctx, "TestValue", "through_context")
	L(ctx).Info("hello, world")
	assert.Equal(t, `{"time":"","level":"INFO","msg":"hello, world","TestValue":"through_context"}`, sink.Get())
}
