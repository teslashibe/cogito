package xlog

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"strings"
)

var logger *slog.Logger
var addSource bool

func init() {
	ConfigureFromEnv()
}

func _log(level slog.Level, msg string, args ...any) {
	if addSource {
		_, f, l, _ := runtime.Caller(2)
		group := slog.Group(
			"source",
			slog.Attr{
				Key:   "file",
				Value: slog.AnyValue(f),
			},
			slog.Attr{
				Key:   "L",
				Value: slog.AnyValue(l),
			},
		)
		args = append(args, group)
	}
	logger.Log(context.Background(), level, msg, args...)
}

func Info(msg string, args ...any) {
	_log(slog.LevelInfo, msg, args...)
}

func Debug(msg string, args ...any) {
	_log(slog.LevelDebug, msg, args...)
}

func Error(msg string, args ...any) {
	_log(slog.LevelError, msg, args...)
}

func Warn(msg string, args ...any) {
	_log(slog.LevelWarn, msg, args...)
}

// ConfigureFromEnv configures logging using LOG_LEVEL, LOG_FORMAT, LOG_ADD_SOURCE.
func ConfigureFromEnv() {
	level := strings.ToLower(strings.TrimSpace(os.Getenv("LOG_LEVEL")))
	format := strings.ToLower(strings.TrimSpace(os.Getenv("LOG_FORMAT")))
	addSourceEnv := strings.ToLower(strings.TrimSpace(os.Getenv("LOG_ADD_SOURCE")))
	Configure(level, format, addSourceEnv == "true")
}

// Configure reconfigures the logger at runtime.
func Configure(level, format string, withSource bool) {
	addSource = withSource
	handler := newHandler(level, format)
	logger = slog.New(handler)
}

func newHandler(level, format string) slog.Handler {
	lvl := slog.LevelDebug
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "info":
		lvl = slog.LevelInfo
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	case "debug":
		lvl = slog.LevelDebug
	}

	opts := &slog.HandlerOptions{
		Level: lvl,
	}

	switch strings.ToLower(strings.TrimSpace(format)) {
	case "json":
		return slog.NewJSONHandler(os.Stdout, opts)
	case "pretty":
		return &prettyHandler{level: lvl}
	default:
		return slog.NewTextHandler(os.Stdout, opts)
	}
}

// prettyHandler outputs logs in a clean, human-readable format.
type prettyHandler struct {
	level slog.Level
	attrs []slog.Attr
	group string
}

func (h *prettyHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *prettyHandler) Handle(_ context.Context, r slog.Record) error {
	timeStr := r.Time.Format("15:04:05")
	levelStr := strings.ToUpper(r.Level.String())

	// Pad level to 5 chars for alignment
	for len(levelStr) < 5 {
		levelStr += " "
	}

	// Build attributes string
	var attrs strings.Builder
	r.Attrs(func(a slog.Attr) bool {
		if a.Key != "" && a.Key != "source" {
			if attrs.Len() > 0 {
				attrs.WriteString(" ")
			}
			attrs.WriteString(a.Key)
			attrs.WriteString("=")
			attrs.WriteString(a.Value.String())
		}
		return true
	})

	if attrs.Len() > 0 {
		fmt.Fprintf(os.Stdout, "%s %s %s %s\n", timeStr, levelStr, r.Message, attrs.String())
	} else {
		fmt.Fprintf(os.Stdout, "%s %s %s\n", timeStr, levelStr, r.Message)
	}
	return nil
}

func (h *prettyHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &prettyHandler{level: h.level, attrs: append(h.attrs, attrs...), group: h.group}
}

func (h *prettyHandler) WithGroup(name string) slog.Handler {
	return &prettyHandler{level: h.level, attrs: h.attrs, group: name}
}
