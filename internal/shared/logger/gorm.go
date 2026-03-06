package logger

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	gormlogger "gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

const defaultSlowSQLThreshold = 200 * time.Millisecond

var sqlKeywordPattern = regexp.MustCompile(`(?i)\b(select|insert|into|values|update|set|delete|from|where|join|left|right|inner|outer|on|order|by|group|limit|offset|returning|and|or|in|between|asc|desc)\b`)

type slogGORMLogger struct {
	logger                    *slog.Logger
	logLevel                  gormlogger.LogLevel
	slowThreshold             time.Duration
	ignoreRecordNotFoundError bool
}

type devGORMLogger struct {
	w                         io.Writer
	logLevel                  gormlogger.LogLevel
	slowThreshold             time.Duration
	ignoreRecordNotFoundError bool
}

func NewGORMLogger(env, level string) gormlogger.Interface {
	normalizedEnv := normalizeEnv(env)
	logLevel := parseGORMLogLevel(normalizedEnv, level)

	if normalizedEnv != envProduction {
		return newDevGORMLogger(os.Stdout, logLevel)
	}

	return &slogGORMLogger{
		logger:                    Default().With("component", "db"),
		logLevel:                  logLevel,
		slowThreshold:             defaultSlowSQLThreshold,
		ignoreRecordNotFoundError: true,
	}
}

func newDevGORMLogger(w io.Writer, logLevel gormlogger.LogLevel) gormlogger.Interface {
	return &devGORMLogger{
		w:                         w,
		logLevel:                  logLevel,
		slowThreshold:             defaultSlowSQLThreshold,
		ignoreRecordNotFoundError: true,
	}
}

func colorizeSQL(line string) string {
	return sqlKeywordPattern.ReplaceAllStringFunc(line, func(match string) string {
		return Cyan + strings.ToUpper(match) + Reset
	})
}

func (l *devGORMLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	next := *l
	next.logLevel = level
	return &next
}

func (l *devGORMLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel < gormlogger.Info {
		return
	}

	l.writeLine(ctx, fmt.Sprintf("%s[gorm:info]%s %s", Cyan, Reset, formatGORMMessage(msg, data...)))
}

func (l *devGORMLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel < gormlogger.Warn {
		return
	}

	l.writeLine(ctx, fmt.Sprintf("%s[gorm:warn]%s %s", Yellow, Reset, formatGORMMessage(msg, data...)))
}

func (l *devGORMLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel < gormlogger.Error {
		return
	}

	l.writeLine(ctx, fmt.Sprintf("%s[gorm:error]%s %s", Red, Reset, formatGORMMessage(msg, data...)))
}

func (l *devGORMLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.logLevel <= gormlogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()
	rowsText := formatGORMRows(rows)
	line := fmt.Sprintf("[sql %.3fms] [rows:%s] %s", float64(elapsed.Nanoseconds())/1e6, rowsText, colorizeSQL(sql))

	switch {
	case err != nil && l.logLevel >= gormlogger.Error && (!errors.Is(err, gormlogger.ErrRecordNotFound) || !l.ignoreRecordNotFoundError):
		l.writeLine(ctx, fmt.Sprintf("%s[sql:error %.3fms]%s [rows:%s] %s err=%v",
			Red,
			float64(elapsed.Nanoseconds())/1e6,
			Reset,
			rowsText,
			colorizeSQL(sql),
			err,
		))
	case l.slowThreshold > 0 && elapsed > l.slowThreshold && l.logLevel >= gormlogger.Warn:
		l.writeLine(ctx, fmt.Sprintf("%s[sql:slow %.3fms]%s [rows:%s] %s",
			Yellow,
			float64(elapsed.Nanoseconds())/1e6,
			Reset,
			rowsText,
			colorizeSQL(sql),
		))
	case l.logLevel >= gormlogger.Info:
		l.writeLine(ctx, line)
	}
}

func (l *devGORMLogger) writeLine(ctx context.Context, line string) {
	if AppendRequestEvent(ctx, line) {
		return
	}

	prefix := ScopePrefix(ctx)
	if prefix != "" {
		line = prefix + " " + line
	}
	_, _ = fmt.Fprintln(l.w, line)
}

func (l *slogGORMLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	next := *l
	next.logLevel = level
	return &next
}

func (l *slogGORMLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel >= gormlogger.Info {
		attrs := append([]any{}, ScopeAttrs(ctx)...)
		attrs = append(attrs, "source", utils.FileWithLineNum(), "message", msg)
		attrs = append(attrs, data...)
		l.logger.Info("gorm info", attrs...)
	}
}

func (l *slogGORMLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel >= gormlogger.Warn {
		attrs := append([]any{}, ScopeAttrs(ctx)...)
		attrs = append(attrs, "source", utils.FileWithLineNum(), "message", msg)
		attrs = append(attrs, data...)
		l.logger.Warn("gorm warn", attrs...)
	}
}

func (l *slogGORMLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel >= gormlogger.Error {
		attrs := append([]any{}, ScopeAttrs(ctx)...)
		attrs = append(attrs, "source", utils.FileWithLineNum(), "message", msg)
		attrs = append(attrs, data...)
		l.logger.Error("gorm error", attrs...)
	}
}

func (l *slogGORMLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.logLevel <= gormlogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	attrs := append([]any{}, ScopeAttrs(ctx)...)
	attrs = append(attrs,
		"source", utils.FileWithLineNum(),
		"elapsed_ms", float64(elapsed.Nanoseconds())/1e6,
		"sql", sql,
	)

	if rows >= 0 {
		attrs = append(attrs, "rows", rows)
	}

	switch {
	case err != nil && l.logLevel >= gormlogger.Error && (!errors.Is(err, gormlogger.ErrRecordNotFound) || !l.ignoreRecordNotFoundError):
		l.logger.Error("gorm query error", append(attrs, "error", err.Error())...)
	case l.slowThreshold > 0 && elapsed > l.slowThreshold && l.logLevel >= gormlogger.Warn:
		l.logger.Warn("gorm slow query", append(attrs, "threshold_ms", l.slowThreshold.Milliseconds())...)
	case l.logLevel >= gormlogger.Info:
		l.logger.Info("gorm query", attrs...)
	}
}

func formatGORMRows(rows int64) string {
	if rows < 0 {
		return "-"
	}

	return strconv.FormatInt(rows, 10)
}

func formatGORMMessage(msg string, data ...interface{}) string {
	if len(data) == 0 {
		return msg
	}

	return fmt.Sprintf(msg, data...)
}

func parseGORMLogLevel(env, level string) gormlogger.LogLevel {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "silent":
		return gormlogger.Silent
	case "error":
		return gormlogger.Error
	case "warn", "warning":
		return gormlogger.Warn
	case "info":
		return gormlogger.Info
	case "":
		if normalizeEnv(env) == envProduction {
			return gormlogger.Warn
		}

		return gormlogger.Info
	default:
		if normalizeEnv(env) == envProduction {
			return gormlogger.Warn
		}

		return gormlogger.Info
	}
}
