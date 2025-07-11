package mqtt

import "log/slog"

func (l LogLevel) toSlogLevel() slog.Level {
	switch l {
	case LogLevelDebug:
		return slog.LevelDebug
	case LogLevelWarn:
		return slog.LevelWarn
	case LogLevelError:
		return slog.LevelError
	default:
		return slog.LevelError
	}
}

func componentAttr(tag component) slog.Attr {
	return slog.String("component", string(tag))
}
