package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	envDevelopment = "development"
	envProduction  = "production"
	envTest        = "test"
	Reset          = "\033[0m"
	Green          = "\033[32m"
	Blue           = "\033[34m"
	Magenta        = "\033[35m"
	Cyan           = "\033[36m"
	White          = "\033[37m"
	Yellow         = "\033[33m"
	Red            = "\033[31m"
)

var (
	mu        sync.RWMutex
	current   = newLogger(os.Stdout, "", envDevelopment)
	devLogMu  sync.Mutex
	devOutput io.Writer = os.Stdout
)

func Init(appName, env string) {
	logger := newLogger(os.Stdout, appName, env)

	mu.Lock()
	current = logger
	mu.Unlock()

	slog.SetDefault(logger)
}

func Default() *slog.Logger {
	mu.RLock()
	defer mu.RUnlock()

	return current
}

func Debug(msg string, args ...any) {
	Default().Debug(msg, args...)
}

func DebugContext(ctx context.Context, msg string, args ...any) {
	if AppendRequestEvent(ctx, formatContextLogLine("", msg, args...)) {
		return
	}

	Default().Debug(ScopeMessage(ctx, msg), args...)
}

func Info(msg string, args ...any) {
	Default().Info(msg, args...)
}

func InfoContext(ctx context.Context, msg string, args ...any) {
	if AppendRequestEvent(ctx, formatContextLogLine("[info]", msg, args...)) {
		return
	}

	Default().Info(ScopeMessage(ctx, msg), args...)
}

func Warn(msg string, args ...any) {
	Default().Warn(msg, args...)
}

func WarnContext(ctx context.Context, msg string, args ...any) {
	if AppendRequestEvent(ctx, formatContextLogLine(Yellow+"[warn]"+Reset, msg, args...)) {
		return
	}

	Default().Warn(ScopeMessage(ctx, msg), args...)
}

func Error(msg string, args ...any) {
	Default().Error(msg, args...)
}

func ErrorContext(ctx context.Context, msg string, args ...any) {
	if AppendRequestEvent(ctx, formatContextLogLine(Red+"[error]"+Reset, msg, args...)) {
		return
	}

	Default().Error(ScopeMessage(ctx, msg), args...)
}

func newLogger(w io.Writer, appName, env string) *slog.Logger {
	normalizedEnv := normalizeEnv(env)
	handler := newHandler(w, normalizedEnv)
	logger := slog.New(handler)

	if normalizedEnv == envProduction && appName != "" {
		logger = logger.With("app", appName)
	}

	if normalizedEnv == envProduction {
		logger = logger.With("env", normalizedEnv)
	}

	return logger
}

func newHandler(w io.Writer, env string) slog.Handler {
	if env == envProduction {
		return slog.NewJSONHandler(w, &slog.HandlerOptions{
			Level: slog.LevelInfo,
			ReplaceAttr: func(_ []string, attr slog.Attr) slog.Attr {
				if attr.Key == slog.TimeKey {
					attr.Value = slog.StringValue(attr.Value.Time().UTC().Format(time.RFC3339Nano))
				}

				return attr
			},
		})
	}

	return slog.NewTextHandler(ansiWriter{w: w}, &slog.HandlerOptions{
		Level: slog.LevelDebug,
		ReplaceAttr: func(_ []string, attr slog.Attr) slog.Attr {
			switch attr.Key {
			case slog.TimeKey:
				attr.Value = slog.StringValue(attr.Value.Time().Format("15:04:05.000"))
			}

			return attr
		},
	})
}

func normalizeEnv(env string) string {
	switch strings.ToLower(strings.TrimSpace(env)) {
	case "", "dev", "development", "local":
		return envDevelopment
	case "prod", "production":
		return envProduction
	case "test", "testing":
		return envTest
	default:
		return strings.ToLower(strings.TrimSpace(env))
	}
}

func formatContextLogLine(levelPrefix, msg string, args ...any) string {
	line := strings.TrimSpace(msg)
	if levelPrefix != "" {
		line = strings.TrimSpace(levelPrefix + " " + line)
	}

	if len(args) == 0 {
		return line
	}

	var attrs []string
	for i := 0; i < len(args); i += 2 {
		key := fmt.Sprint(args[i])
		if i+1 >= len(args) {
			attrs = append(attrs, key)
			continue
		}

		attrs = append(attrs, key+"="+fmt.Sprint(args[i+1]))
	}

	return line + " " + strings.Join(attrs, " ")
}

type ansiWriter struct {
	w io.Writer
}

func (w ansiWriter) Write(p []byte) (int, error) {
	line := string(p)
	line = strings.ReplaceAll(line, "level=DEBUG", "level=\033[36mDEBUG\033[0m")
	line = strings.ReplaceAll(line, "level=INFO", "level=\033[32mINFO\033[0m")
	line = strings.ReplaceAll(line, "level=WARN", "level=\033[33mWARN\033[0m")
	line = strings.ReplaceAll(line, "level=ERROR", "level=\033[31mERROR\033[0m")

	if _, err := io.WriteString(w.w, line); err != nil {
		return 0, err
	}

	return len(p), nil
}
