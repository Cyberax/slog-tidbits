package tidbits

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"strings"
	"testing"
)

func TestPrettySink(t *testing.T) {
	t.Parallel()

	data := &bytes.Buffer{}
	pretty := NewPrettySink(data, slog.LevelDebug, true)

	conv := slog.New(pretty.GetHandler())
	conv.Debug("hello, world", slog.Int64("key", 42), slog.Int64("key2", 1<<55))
	conv.Info("Info Message")
	conv.Warn("Warning Message")
	conv.Error("Error Message")

	expected := `[37mDEBUG[0m  pretty_sink_test.go:18  hello, world  key=42  key2=36028797018963968  
INFO  pretty_sink_test.go:19  Info Message  
[33mWARN[0m  pretty_sink_test.go:20  Warning Message  
[31mERROR[0m  pretty_sink_test.go:21  Error Message`
	assert.Equal(t, expected, removeTimes(data.String()))
}

func TestPrettySinkStacks(t *testing.T) {
	t.Parallel()

	data := &bytes.Buffer{}
	pretty := NewPrettySink(data, slog.LevelInfo, true)

	conv := slog.New(pretty.GetHandler())
	conv.Error("Happened", StackTraceAttr(false, "it's exploding"),
		slog.Int64("key", 42))

	expected := `[31mERROR[0m  pretty_sink_test.go:37  Happened  key=42
	panic: it's exploding
	github.com/Cyberax/slog-tidbits/tidbits/stacks.go:27 (StackTraceAttr)
	github.com/Cyberax/slog-tidbits/tidbits/pretty_sink_test.go:37 (TestPrettySinkStacks)
	testing/testing.go:1689 (tRunner)`

	assert.Equal(t, expected, removeTimes(data.String()))
}

func removeTimes(logs string) string {
	res := ""
	for _, ln := range strings.Split(logs, "\n") {
		parts := strings.SplitN(ln, "  ", 2)
		if len(parts) == 2 {
			res += parts[1] + "\n"
		} else {
			res += ln + "\n"
		}
	}
	return strings.TrimSpace(res)
}
