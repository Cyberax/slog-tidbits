package tidbits

import "log/slog"

type ControlAttr interface {
	slog.LogValuer
}

type AttrOrder struct {
	appendRight bool
}

var _ ControlAttr = &AttrOrder{}

type AttrLevel struct {
	level slog.Level
}

var _ ControlAttr = &AttrLevel{}

func (c *AttrOrder) LogValue() slog.Value {
	if c.appendRight {
		return slog.StringValue("right")
	} else {
		return slog.StringValue("left")
	}
}

func AddToLeft() slog.Attr {
	return slog.Any("set_order", &AttrOrder{appendRight: false})
}

func AddToRight() slog.Attr {

	return slog.Any("set_order", &AttrOrder{appendRight: true})
}

func (c *AttrLevel) LogValue() slog.Value {
	return slog.StringValue(c.level.String())
}

func WithLogLevel(lvl slog.Level) slog.Attr {
	return slog.Any("set_level", &AttrLevel{level: lvl})
}
