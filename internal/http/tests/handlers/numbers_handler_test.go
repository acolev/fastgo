package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"fastgo/internal/i18n"
	"fastgo/internal/shared/response"

	"github.com/gofiber/fiber/v3"
)

func TestCreateRangeInvalidBodyReturnsLocalizedJSONError(t *testing.T) {
	i18n.SetDefaultLang("en")
	if err := i18n.LoadDir(filepath.Join("..", "..", "..", "..", "locales")); err != nil {
		t.Fatalf("load locales: %v", err)
	}

	app := fiber.New(fiber.Config{
		ErrorHandler: response.ErrorHandler,
	})
	app.Use(i18n.Middleware())

	handler := New()
	app.Post("/numbers/range", handler.CreateRange)

	req := httptest.NewRequest(http.MethodPost, "/numbers/range", strings.NewReader("{"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Lang", "ru")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app test: %v", err)
	}
	t.Cleanup(func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			t.Fatalf("close body: %v", closeErr)
		}
	})

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("got status %d, want %d", resp.StatusCode, fiber.StatusBadRequest)
	}

	var payload response.ErrorEnvelope
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode body: %v", err)
	}

	if payload.Error.Code != "invalid_request_body" {
		t.Fatalf("got code %q, want %q", payload.Error.Code, "invalid_request_body")
	}

	if payload.Error.Message != "Некорректное JSON-тело запроса" {
		t.Fatalf("got message %q, want %q", payload.Error.Message, "Некорректное JSON-тело запроса")
	}
}
