package tidbits

import (
	"github.com/stretchr/testify/assert"
	"log/slog"
	"testing"
)

func TestControlsAttrsRendering(t *testing.T) {
	t.Parallel()
	sl := NewSinkingLogger(slog.LevelInfo)

	val := AddToLeft()
	sl.Info("Hello", val)
	val2 := AddToRight()
	sl.Info("Hello", val2)
	val3 := WithLogLevel(slog.LevelError)
	sl.Info("Hello", val3)
	assert.Equal(t, `{"time":"","level":"INFO","msg":"Hello","set_order":"left"}
{"time":"","level":"INFO","msg":"Hello","set_order":"right"}
{"time":"","level":"INFO","msg":"Hello","set_level":"ERROR"}`, sl.Get())
}
