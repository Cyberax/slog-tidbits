package tidbits

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slog"
	"runtime"
	"strings"
	"testing"
)

func TestPackageOverrides(t *testing.T) {
	lvls := NewPinpointLogLevels()

	lvls.WithEnvironmentListOverrides([]string{
		"TEST=asd",
		"TIDBITS_LOG_WARN=github.com/package1,github.com/package2",
		"TIDBITS_LOG_ERROR=github.com/package1/subpackage2",
		"TIDBITS_LOG_ERROR_MINUS_8=github.com/package3",
		"TIDBITS_LOG_DEBUG_PLUS_4=github.com/pack",
		"TIDBITS_LOG_WARN_PLUS_4=",
	})

	l, ok := lvls.LevelForPackage("github.com/pack")
	assert.True(t, ok)
	assert.Equal(t, slog.LevelInfo, l)

	l, ok = lvls.LevelForPackage("github.com/package1/subpackage2")
	assert.Equal(t, slog.LevelError, l)

	l, ok = lvls.LevelForPackage("github.com/package1")
	assert.Equal(t, slog.LevelWarn, l)

	_, ok = lvls.LevelForPackage("github.com/something")
	assert.False(t, ok)
}

func f1() {
	pc, _, _, _ := runtime.Caller(1)
	forPC := runtime.FuncForPC(pc)
	funcName := forPC.Name()
	lastSlash := strings.LastIndexByte(funcName, '/')
	if lastSlash < 0 {
		lastSlash = 0
	}

	lastDot := strings.LastIndexByte(funcName[lastSlash:], '.') + lastSlash

	fmt.Printf("Package: %s\n", funcName[:lastDot])
	fmt.Printf("Func:   %s\n", funcName[lastDot+1:])
}
