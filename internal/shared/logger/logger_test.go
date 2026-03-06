package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	fiberlogger "github.com/gofiber/fiber/v3/middleware/logger"
	gormlogger "gorm.io/gorm/logger"
)

func TestNewLoggerProductionWritesJSON(t *testing.T) {
	var buf bytes.Buffer

	logger := newLogger(&buf, "FastGo", "production")
	logger.Info("started", "addr", ":3005")

	line := strings.TrimSpace(buf.String())
	if line == "" {
		t.Fatal("expected log output")
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(line), &payload); err != nil {
		t.Fatalf("unmarshal log line: %v", err)
	}

	if payload["msg"] != "started" {
		t.Fatalf("msg = %v", payload["msg"])
	}

	if payload["app"] != "FastGo" {
		t.Fatalf("app = %v", payload["app"])
	}

	if payload["env"] != "production" {
		t.Fatalf("env = %v", payload["env"])
	}
}

func TestNewLoggerDevelopmentWritesColoredText(t *testing.T) {
	var buf bytes.Buffer

	logger := newLogger(&buf, "FastGo", "development")
	logger.Info("started", "addr", ":3005")

	output := buf.String()
	if !strings.Contains(output, "\033[32mINFO\033[0m") {
		t.Fatalf("expected colored INFO level, got %q", output)
	}

	if !strings.Contains(output, "msg=started") {
		t.Fatalf("expected text output, got %q", output)
	}

	if strings.Contains(output, "app=FastGo") || strings.Contains(output, "env=development") {
		t.Fatalf("development log should stay compact, got %q", output)
	}
}

func TestParseGORMLogLevelDefaultsByEnv(t *testing.T) {
	if got := parseGORMLogLevel("development", ""); got != gormlogger.Info {
		t.Fatalf("development default = %v, want %v", got, gormlogger.Info)
	}

	if got := parseGORMLogLevel("production", ""); got != gormlogger.Warn {
		t.Fatalf("production default = %v, want %v", got, gormlogger.Warn)
	}

	if got := parseGORMLogLevel("production", "error"); got != gormlogger.Error {
		t.Fatalf("explicit level = %v, want %v", got, gormlogger.Error)
	}
}

func TestHTTPMiddlewareUsesFiberFormatsByEnv(t *testing.T) {
	devHandler := HTTPMiddleware("development")
	if devHandler == nil {
		t.Fatal("development middleware is nil")
	}

	prodHandler := HTTPMiddleware("production")
	if prodHandler == nil {
		t.Fatal("production middleware is nil")
	}

	_ = fiberlogger.DefaultFormat
	_ = fiberlogger.JSONFormat
}

func TestColorizeSQLHighlightsKeywords(t *testing.T) {
	line := `SELECT * FROM "numbers" WHERE "number" = 1 ORDER BY "number" ASC`
	colored := colorizeSQL(line)

	if !strings.Contains(colored, Cyan+"SELECT"+Reset) {
		t.Fatalf("expected SELECT highlight, got %q", colored)
	}

	if !strings.Contains(colored, Cyan+"FROM"+Reset) {
		t.Fatalf("expected FROM highlight, got %q", colored)
	}

	if !strings.Contains(colored, Cyan+"WHERE"+Reset) {
		t.Fatalf("expected WHERE highlight, got %q", colored)
	}
}

func TestScopeMessageIncludesRequestScope(t *testing.T) {
	ctx := WithRequestScope(context.Background(), RequestScope{
		Method: "GET",
		Path:   "/api/v1/t/numbers",
	})

	got := ScopeMessage(ctx, "[cache:hit] tests:numbers:list")
	want := "[GET /api/v1/t/numbers] [cache:hit] tests:numbers:list"
	if got != want {
		t.Fatalf("scope message = %q, want %q", got, want)
	}
}

func TestDevGORMLoggerTraceIncludesScopePrefix(t *testing.T) {
	var buf bytes.Buffer

	logger := newDevGORMLogger(&buf, gormlogger.Info)
	ctx := WithRequestScope(context.Background(), RequestScope{
		Method: "GET",
		Path:   "/api/v1/t/numbers",
	})

	logger.Trace(ctx, time.Now().Add(-1500*time.Microsecond), func() (string, int64) {
		return `SELECT * FROM "numbers" ORDER BY "number" ASC`, 100
	}, nil)

	output := buf.String()
	if !strings.Contains(output, "[GET /api/v1/t/numbers] [sql ") {
		t.Fatalf("expected scoped SQL prefix, got %q", output)
	}

	if !strings.Contains(output, "[rows:100]") {
		t.Fatalf("expected rows count, got %q", output)
	}

	if !strings.Contains(output, Cyan+"SELECT"+Reset) {
		t.Fatalf("expected highlighted SELECT, got %q", output)
	}
}

func TestPrintDevRequestBlockKeepsFiberHeader(t *testing.T) {
	var buf bytes.Buffer

	prev := devOutput
	devOutput = &buf
	t.Cleanup(func() {
		devOutput = prev
	})

	header := "22:05:44 |\x1b[32m 200 \x1b[0m|   80.136083ms |       127.0.0.1 |\x1b[32m POST   \x1b[0m| /v2/auth/refresh -"
	printDevRequestBlock(header, []string{"[cache:miss] tests:numbers:list"})

	output := buf.String()
	if !strings.Contains(output, "┌ "+header) {
		t.Fatalf("expected original fiber header in block, got %q", output)
	}

	if !strings.Contains(output, "│ [cache:miss] tests:numbers:list") {
		t.Fatalf("expected grouped event line, got %q", output)
	}
}
