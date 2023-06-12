package tidbits

import (
	"github.com/stretchr/testify/assert"
	"log/slog"
	"testing"
)

func TestAppendDirection(t *testing.T) {
	t.Parallel()

	sink := NewSinkingLogger(slog.LevelInfo)

	conv := slog.New(NewSlogConvenience(SlogOptions{}, sink.Handler()))

	conv = conv.With(slog.String("initial", "position"))
	d1 := conv.With(AddToRight()).With(slog.String("first", "right")).
		With(AddToLeft()).With(slog.String("second", "left"))
	d1.Info("hello, world")

	d1.Info("hello, world", "prepended", "value")

	res := sink.Get()
	assert.Equal(t, `{"time":"","level":"INFO","msg":"hello, world","second":"left","initial":"position","first":"right"}
{"time":"","level":"INFO","msg":"hello, world","prepended":"value","second":"left","initial":"position","first":"right"}`, res)
}

func TestLevelOverride(t *testing.T) {
	t.Parallel()

	sink := NewSinkingLogger(slog.LevelDebug)
	conv := slog.New(NewSlogConvenience(SlogOptions{
		LogLevel: slog.LevelWarn,
	}, sink.Handler()))

	conv.Info("hello, world")
	assert.Empty(t, sink.Get())

	conv = conv.With(WithLogLevel(slog.LevelInfo))
	conv.Info("hello, world")
	assert.Equal(t, `{"time":"","level":"INFO","msg":"hello, world"}`, sink.Get())
}

func TestPinpointer(t *testing.T) {
	t.Parallel()

	sink := NewSinkingLogger(slog.LevelDebug)
	levels := NewPinpointLogLevels()
	levels.WithOverride(slog.LevelError, "github.com/Cyberax/slog-tidbits")
	levels.WithOverride(slog.LevelInfo, "github.com/Cyberax/slog-tidbits/tidbits.interesting")

	conv := slog.New(NewSlogConvenience(SlogOptions{
		Pinpointer: levels,
	}, sink.Handler()))

	conv.Info("hello, world")
	assert.Empty(t, sink.Get())

	interesting(conv)
	assert.Equal(t, `{"time":"","level":"INFO","msg":"interesting message"}`, sink.Get())
}

func interesting(log *slog.Logger) {
	log.Info("interesting message")
}
