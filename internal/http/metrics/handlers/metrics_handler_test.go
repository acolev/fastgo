package handlers

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
)

func TestMetricsHandlerRespondsWithPrometheusPayload(t *testing.T) {
	app := fiber.New()
	app.Get("/metrics", New())

	req := httptest.NewRequest("GET", "/metrics", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}

	if got := resp.Header.Get(fiber.HeaderContentType); !strings.Contains(got, "text/plain") {
		t.Fatalf("content-type = %q", got)
	}
}
