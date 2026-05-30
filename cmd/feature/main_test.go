package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateCreatesFeature(t *testing.T) {
	root := t.TempDir()
	if err := generate(root, "users"); err != nil {
		t.Fatalf("generate returned error: %v", err)
	}

	expectedFiles := []string{
		"internal/http/users/dto/users_dto.go",
		"internal/http/users/handlers/users_handler.go",
		"internal/http/users/services/service.go",
		"internal/http/users/routes.go",
	}

	for _, path := range expectedFiles {
		if _, err := os.Stat(filepath.Join(root, path)); err != nil {
			t.Fatalf("expected generated file %s: %v", path, err)
		}
	}

	routes, err := os.ReadFile(filepath.Join(root, "internal/http/users/routes.go"))
	if err != nil {
		t.Fatalf("read generated routes: %v", err)
	}

	if !strings.Contains(string(routes), `usersGroup := r.Group("/users")`) {
		t.Fatalf("expected generated feature group, got:\n%s", routes)
	}

	if !strings.Contains(string(routes), `usersGroup.Get("", handler.List)`) {
		t.Fatalf("expected generated list route, got:\n%s", routes)
	}

	handler, err := os.ReadFile(filepath.Join(root, "internal/http/users/handlers/users_handler.go"))
	if err != nil {
		t.Fatalf("read generated handler: %v", err)
	}

	if !strings.Contains(string(handler), "func (h *Handler) List") {
		t.Fatalf("expected generated list handler, got:\n%s", handler)
	}
}

func TestGenerateRejectsInvalidName(t *testing.T) {
	if err := generate(t.TempDir(), "../users"); err == nil {
		t.Fatal("expected invalid name error")
	}
}

func TestGenerateDoesNotOverwriteFeature(t *testing.T) {
	root := t.TempDir()
	if err := generate(root, "users"); err != nil {
		t.Fatalf("first generate returned error: %v", err)
	}

	if err := generate(root, "users"); err == nil {
		t.Fatal("expected existing feature error")
	}
}
