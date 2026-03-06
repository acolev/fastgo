package logger

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/gofiber/fiber/v3"
	fiberlogger "github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/requestid"
)

const prodHTTPLogFormat = "{\"time\":\"${time}\",\"status\":${status},\"latency\":\"${latency}\",\"method\":\"${method}\",\"path\":\"${path}\",\"request_id\":\"${reqHeader:X-Request-ID}\",\"error\":\"${error}\"}\n"

func HTTPMiddleware(env string) fiber.Handler {
	env = normalizeEnv(env)

	if env == envProduction {
		return fiberlogger.New(fiberlogger.Config{
			DisableColors: true,
			Format:        prodHTTPLogFormat,
			Skip:          shouldSkipHTTPLog,
			TimeFormat:    "2006-01-02T15:04:05.000Z07:00",
			TimeZone:      "UTC",
		})
	}

	httpLogger := fiberlogger.New(fiberlogger.Config{
		Format:      fiberlogger.DefaultFormat,
		Skip:        shouldSkipHTTPLog,
		TimeFormat:  "15:04:05",
		TimeZone:    "Local",
		ForceColors: true,
		Stream:      io.Discard,
		Done: func(c fiber.Ctx, logString []byte) {
			printDevRequestBlock(strings.TrimRight(string(logString), "\n"), DrainRequestEvents(c.Context()))
		},
	})

	return newDevHTTPMiddleware(httpLogger)
}

func newDevHTTPMiddleware(httpLogger fiber.Handler) fiber.Handler {
	return func(c fiber.Ctx) error {
		baseCtx := c.Context()
		if baseCtx == nil {
			baseCtx = context.Background()
		}

		scope := RequestScope{
			Method:    c.Method(),
			Path:      c.Path(),
			IP:        c.IP(),
			RequestID: requestid.FromContext(c),
		}
		ctx := WithRequestScope(baseCtx, scope)
		if !shouldSkipHTTPLog(c) {
			ctx = WithRequestLogBuffer(ctx)
		}
		c.SetContext(ctx)

		return httpLogger(c)
	}
}

func shouldSkipHTTPLog(c fiber.Ctx) bool {
	path := c.Path()

	switch path {
	case "/ping", "/health", "/ready":
		return true
	}

	return strings.HasPrefix(path, "/docs") || strings.HasPrefix(path, "/api/docs")
}

func printDevRequestBlock(header string, events []string) {
	devLogMu.Lock()
	defer devLogMu.Unlock()

	_, _ = fmt.Fprintln(devOutput)
	_, _ = fmt.Fprintln(devOutput, "┌ "+header)
	for _, event := range events {
		if strings.TrimSpace(event) == "" {
			continue
		}

		_, _ = fmt.Fprintln(devOutput, "│ "+event)
	}
	_, _ = fmt.Fprintln(devOutput, "└")
}
