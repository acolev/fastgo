package response_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"fastgo/internal/i18n"
	"fastgo/internal/shared/response"

	"github.com/gofiber/fiber/v3"
)

func TestErrorHandlerReturnsJSONForNotFound(t *testing.T) {
	i18n.SetDefaultLang("en")
	if err := i18n.LoadDir(filepath.Join("..", "..", "..", "locales")); err != nil {
		t.Fatalf("load locales: %v", err)
	}

	app := fiber.New(fiber.Config{
		ErrorHandler: response.ErrorHandler,
	})
	app.Use(i18n.Middleware())

	req := httptest.NewRequest(http.MethodGet, "/missing", nil)
	req.Header.Set("X-Lang", "ru")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusNotFound {
		t.Fatalf("got status %d, want %d", resp.StatusCode, fiber.StatusNotFound)
	}

	var payload response.ErrorEnvelope
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode body: %v", err)
	}

	if payload.Error.Code != "not_found" {
		t.Fatalf("got code %q, want %q", payload.Error.Code, "not_found")
	}

	if payload.Error.Message != "Ресурс не найден" {
		t.Fatalf("got message %q, want %q", payload.Error.Message, "Ресурс не найден")
	}
}
