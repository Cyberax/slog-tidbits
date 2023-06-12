package tidbits

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
)

// SinkingLogger This is a logger that writes messages to a memory buffer, it's intended to be
// used in unit tests to make sure that the log messages are expected. As a convenience feature
// to make it easier to match the messages, this logger also removes the timestamp from the log
// entries.
type SinkingLogger struct {
	*slog.Logger
	data *bytes.Buffer
}

func NewSinkingLogger(lvl slog.Level) *SinkingLogger {
	return NewJsonOrTextSinkingLogger(lvl, false)
}

func NewJsonOrTextSinkingLogger(lvl slog.Level, textMode bool) *SinkingLogger {
	data := &bytes.Buffer{}
	opts := &slog.HandlerOptions{
		Level:     lvl,
		AddSource: false,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Remove the time from the log messages
			if a.Key == slog.TimeKey {
				return slog.String(slog.TimeKey, "")
			}
			return a
		},
	}

	var lh slog.Handler
	if textMode {
		lh = slog.NewTextHandler(data, opts)
	} else {
		lh = slog.NewJSONHandler(data, opts)
	}

	return &SinkingLogger{
		Logger: slog.New(lh),
		data:   data,
	}
}

// Get returns the accumulated log data and resets the buffer
func (s *SinkingLogger) Get() string {
	res := s.data.String()
	s.data.Reset()
	return strings.TrimSpace(res)
}

// NopLogger This is a logger that does nothing, it's intended to be used in unit tests to
// suppress log messages.
type NopLogger struct {
	*slog.Logger
}

func NewNopLogger(lvl slog.Level) *NopLogger {
	return &NopLogger{
		Logger: slog.New(&nopHandler{}),
	}
}

type nopHandler struct {
}

var _ slog.Handler = &nopHandler{}

func (n *nopHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return false
}

func (n *nopHandler) Handle(ctx context.Context, record slog.Record) error {
	return nil
}

func (n *nopHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return n
}

func (n *nopHandler) WithGroup(name string) slog.Handler {
	return n
}
