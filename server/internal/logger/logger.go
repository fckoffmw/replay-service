package logger

import (
	"log/slog"
	"os"
	"strings"
)

const (
	DEBUG = "debug"
	INFO  = "info"
	WARN  = "warn"
	ERROR = "error"
)

func NewSlog(level string) *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slogLevelFromStr(level),
	})
	return slog.New(handler)
}

func slogLevelFromStr(level string) slog.Leveler {
	var slogLevel slog.Leveler

	switch strings.ToLower(level) {
	case INFO:
		slogLevel = slog.LevelInfo
	case WARN:
		slogLevel = slog.LevelWarn
	case ERROR:
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelDebug
	}

	return slogLevel
}
