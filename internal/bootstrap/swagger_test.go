package bootstrap

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	fiberswagger "github.com/gofiber/contrib/v3/swagger"
	"github.com/gofiber/fiber/v3"
)

func TestLoadSwaggerSpecOverridesTitle(t *testing.T) {
	dir := t.TempDir()
	specPath := filepath.Join(dir, "swagger.json")

	raw := []byte(`{"info":{"title":"Old Name","version":"1.0"}}`)
	if err := os.WriteFile(specPath, raw, 0o644); err != nil {
		t.Fatalf("write spec: %v", err)
	}

	updated, err := loadSwaggerSpec(specPath, "New Name")
	if err != nil {
		t.Fatalf("loadSwaggerSpec: %v", err)
	}

	var spec map[string]any
	if err := json.Unmarshal(updated, &spec); err != nil {
		t.Fatalf("unmarshal updated spec: %v", err)
	}

	info, ok := spec["info"].(map[string]any)
	if !ok {
		t.Fatal("info section missing")
	}

	if got := info["title"]; got != "New Name" {
		t.Fatalf("title = %v", got)
	}
}

func TestSwaggerSpecHandlerOverridesCachedMiddlewareSpec(t *testing.T) {
	app := fiber.New()

	app.Get("/api/docs/swagger.json", swaggerSpecHandler([]byte(`{"info":{"title":"New Name"}}`)))
	app.Use(fiberswagger.New(fiberswagger.Config{
		BasePath:    "/api",
		FilePath:    "docs/swagger.json",
		FileContent: []byte(`{"info":{"title":"Old Name"}}`),
		Path:        "docs",
		Title:       "Docs",
	}))

	req := httptest.NewRequest("GET", "/api/docs/swagger.json", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Fatalf("status = %d", resp.StatusCode)
	}

	if got := resp.Header.Get("Cache-Control"); got != "no-store" {
		t.Fatalf("cache-control = %q", got)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}

	var spec map[string]any
	if err := json.Unmarshal(body, &spec); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}

	info, ok := spec["info"].(map[string]any)
	if !ok {
		t.Fatal("info section missing")
	}

	if got := info["title"]; got != "New Name" {
		t.Fatalf("title = %v", got)
	}
}
