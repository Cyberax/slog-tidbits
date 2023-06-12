package tidbits

import (
	"context"
	"golang.org/x/exp/slog"
)

type SlogConvenience struct {
	delegate slog.Handler

	level slog.Level
	group string
	attrs []slog.Attr
}

var _ slog.Handler = &SlogConvenience{}

func (s *SlogConvenience) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= s.level
}

func (s *SlogConvenience) Handle(ctx context.Context, record slog.Record) error {
	return nil
}

func (s *SlogConvenience) WithAttrs(attrs []slog.Attr) slog.Handler {
	//TODO implement me
	panic("implement me")
}

func (s *SlogConvenience) WithGroup(name string) slog.Handler {
	//TODO implement me
	panic("implement me")
}
