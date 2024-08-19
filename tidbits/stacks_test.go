package tidbits

import (
	"github.com/stretchr/testify/assert"
	"log/slog"
	"strings"
	"testing"
)

func TestStackAttr(t *testing.T) {
	sink := NewSinkingLogger(slog.LevelInfo)
	sink.Error("badmsg", StackTraceAttr(false, "test panic"))

	val := sink.Get()
	// This test is a bit brittle, because the line numbers can change
	expected := `{"time":"","level":"ERROR","msg":"badmsg","stack":[{"panic_msg":"test panic"},
{"fl":"github.com/Cyberax/slog-tidbits/tidbits/stacks.go:27","fn":"StackTraceAttr"},
{"fl":"github.com/Cyberax/slog-tidbits/tidbits/stacks_test.go:12","fn":"TestStackAttr"},
{"fl":"testing/testing.go:1689","fn":"tRunner"}]}`
	assert.Equal(t, strings.ReplaceAll(expected, "\n", ""), val)
}
