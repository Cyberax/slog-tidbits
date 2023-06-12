package tidbits

import (
	"context"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"strings"
	"testing"
)

func TestPackageOverrides(t *testing.T) {
	t.Parallel()

	lvls := NewPinpointLogLevels()

	lvls.WithEnvironmentListOverrides([]string{
		"TEST=asd",
		"TIDBITS_LOG_WARN=github.com/package1,github.com/package2",
		"TIDBITS_LOG_ERROR=github.com/package1/subpackage2",
		"TIDBITS_LOG_ERROR_MINUS_8=github.com/package3",
		"TIDBITS_LOG_DEBUG_PLUS_4=github.com/pack",
		"TIDBITS_LOG_WARN_PLUS_4=",
	})

	l, ok := lvls.LevelForLocation("github.com/pack")
	assert.True(t, ok)
	assert.Equal(t, slog.LevelInfo, l)

	l, ok = lvls.LevelForLocation("github.com/package1/subpackage2")
	assert.Equal(t, slog.LevelError, l)

	l, ok = lvls.LevelForLocation("github.com/package1")
	assert.Equal(t, slog.LevelWarn, l)

	_, ok = lvls.LevelForLocation("github.com/something")
	assert.False(t, ok)

	sl := NewSinkingLogger(slog.LevelInfo)
	lvls.PrintConfig(context.Background(), sl.Logger)

	expected := `{"time":"","level":"INFO","msg":"Effective Pinpoint config","config":[
{"LocationPrefix":"github.com/pack","LogLevel":"INFO"},
{"LocationPrefix":"github.com/package1","LogLevel":"WARN"},
{"LocationPrefix":"github.com/package2","LogLevel":"WARN"},
{"LocationPrefix":"github.com/package3","LogLevel":"INFO"},
{"LocationPrefix":"github.com/package1/subpackage2","LogLevel":"ERROR"}]}`
	assert.Equal(t, strings.ReplaceAll(expected, "\n", ""), sl.Get())
}

func TestLocations(t *testing.T) {
	t.Parallel()

	lvls := NewPinpointLogLevels()

	lvls.WithEnvironmentListOverrides([]string{
		"TIDBITS_LOG_ERROR=github.com/Cyberax/slog-tidbits/tidbits",
		"TIDBITS_LOG_WARN=github.com/Cyberax/slog-tidbits/tidbits.f1",
	})

	f1(t, lvls)

	lvl, _ := lvls.FindLevel(0)
	assert.Equal(t, slog.LevelError, lvl)
}

func f1(t *testing.T, lvls *PinpointLogLevels) {
	lvl, _ := lvls.FindLevel(0)
	assert.Equal(t, slog.LevelWarn, lvl)
}
