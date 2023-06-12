package tidbits

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"log"
	"log/slog"
	"testing"
)

func TestSinkingLog(t *testing.T) {
	sink := NewSinkingLogger(slog.LevelInfo)
	sink.Debug("Missed")
	sink.Info("Hello, world", slog.Any("hello", "world"))
	sink.Error("Test2", slog.Any("test", "123"))

	val := sink.Get()
	expected := `{"time":"","level":"INFO","msg":"Hello, world","hello":"world"}
{"time":"","level":"ERROR","msg":"Test2","test":"123"}`
	assert.Equal(t, expected, val)
}

func TestNop(t *testing.T) {
	old := log.Writer()
	defer log.SetOutput(old)

	out := bytes.NewBuffer(nil)
	log.SetOutput(out)
	sink := NewNopLogger(slog.LevelInfo)
	sink.Error("VERY BAD!")
	assert.True(t, out.Len() == 0)
}
