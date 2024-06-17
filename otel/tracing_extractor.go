package otel

import (
	"context"
	"github.com/Cyberax/slog-tidbits/tidbits"
	"go.opentelemetry.io/otel/trace"
	"log/slog"
)

type TracingIdExtractor struct {
}

var _ tidbits.ContextExtractor = &TracingIdExtractor{}

func (o *TracingIdExtractor) MergeContextAttrs(ctx context.Context, curAttrs []slog.Attr) []slog.Attr {
	span := trace.SpanContextFromContext(ctx)
	if !span.IsValid() {
		return curAttrs
	}
	return append(curAttrs, slog.String("trace_id", span.TraceID().String()),
		slog.String("span_id", span.SpanID().String()))
}
