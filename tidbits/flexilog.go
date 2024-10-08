package tidbits

import (
	"context"
	"log/slog"
	"runtime"
	"slices"
)

type ContextExtractor interface {
	MergeContextAttrs(ctx context.Context, curAttrs []slog.Attr) []slog.Attr
}

type SlogOptions struct {
	AppendNewAttrsRight bool

	Pinpointer *PinpointLogLevels
	LogLevel   slog.Level
	Extractors []ContextExtractor
}

type SlogConvenience struct {
	delegate slog.Handler

	options SlogOptions
	attrs   []slog.Attr
}

var _ slog.Handler = &SlogConvenience{}

func NewSlogConvenience(opts SlogOptions, delegate slog.Handler) *SlogConvenience {
	return &SlogConvenience{
		delegate: delegate,
		options:  opts,
	}
}

func (s *SlogConvenience) Enabled(ctx context.Context, level slog.Level) bool {
	curLevel := s.options.LogLevel
	return level >= curLevel
}

func (s *SlogConvenience) Handle(ctx context.Context, record slog.Record) error {
	// We can't really move this check to Enabled() because it's not really possible to get the
	// calling function name without having access to the slog.Record.PC field.
	if s.options.Pinpointer != nil {
		forPC := runtime.FuncForPC(record.PC)
		funcName := forPC.Name()
		pinpointedLevel, ok := s.options.Pinpointer.LevelForLocation(funcName)
		if ok {
			if record.Level < pinpointedLevel {
				return nil
			}
		}
	}

	newAttrs := make([]slog.Attr, 0, record.NumAttrs())

	var stackTrace *slog.Attr
	record.Attrs(func(a slog.Attr) bool {
		_, ok := a.Value.Any().(*StackValue)
		if stackTrace == nil && a.Key == StackAttrName && ok {
			stackTrace = &a
			return true
		} else {
			newAttrs = append(newAttrs, a)
		}
		return true
	})

	merged := s.mergeAttrs(newAttrs, s.attrs, s.options.AppendNewAttrsRight)

	// Extract context attributes
	for _, extractor := range s.options.Extractors {
		merged = extractor.MergeContextAttrs(ctx, merged)
	}

	mergedRecord := slog.NewRecord(record.Time, record.Level, record.Message, record.PC)
	mergedRecord.AddAttrs(merged...)

	// We don't want a multiline stack trace, but we still want to move it to the end of the message
	if stackTrace != nil {
		mergedRecord.AddAttrs(*stackTrace)
	}

	err := s.delegate.Handle(ctx, mergedRecord)
	if err != nil {
		return err
	}

	return err
}

func (s *SlogConvenience) mergeAttrs(newAttrs, curAttrs []slog.Attr, appendNewAttrsRight bool) []slog.Attr {
	if len(newAttrs) == 0 {
		return slices.Clone(curAttrs)
	}
	if len(curAttrs) == 0 {
		return slices.Clone(newAttrs)
	}

	resAttrs := make([]slog.Attr, 0, len(newAttrs)+len(curAttrs))
	if appendNewAttrsRight {
		resAttrs = append(resAttrs, curAttrs...)
		resAttrs = append(resAttrs, newAttrs...)
	} else {
		resAttrs = append(resAttrs, newAttrs...)
		resAttrs = append(resAttrs, curAttrs...)
	}

	return resAttrs
}

func (s *SlogConvenience) WithAttrs(attrs []slog.Attr) slog.Handler {
	newOptions := s.options

	// Process special-purpose arguments
	attrsProcessed := make([]slog.Attr, 0, len(attrs))
	for _, a := range attrs {
		ca, ok := a.Value.Any().(ControlAttr)
		if ok {
			// This is a control attribute!
			s.processControlAttribute(&newOptions, ca)
			continue
		}
		attrsProcessed = append(attrsProcessed, a)
	}

	resAttrs := s.mergeAttrs(attrsProcessed, s.attrs, newOptions.AppendNewAttrsRight)

	return &SlogConvenience{
		delegate: s.delegate,
		options:  newOptions,
		attrs:    resAttrs,
	}
}

func (s *SlogConvenience) processControlAttribute(opts *SlogOptions, ca ControlAttr) {
	switch v := ca.(type) {
	case *AttrOrder:
		opts.AppendNewAttrsRight = v.appendRight
	case *AttrLevel:
		opts.LogLevel = v.level
	default:
		panic("unknown control attribute")
	}
}

func (s *SlogConvenience) WithGroup(name string) slog.Handler {
	newAttrs := make([]slog.Attr, 0, len(s.attrs))
	for _, a := range s.attrs {
		newAttrs = append(newAttrs, slog.Attr{
			Key:   name + "." + a.Key,
			Value: a.Value,
		})
	}

	return &SlogConvenience{
		delegate: s.delegate,
		options:  s.options,
		attrs:    newAttrs,
	}
}
