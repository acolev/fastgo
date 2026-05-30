package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

var featureNamePattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

func main() {
	if err := run(os.Args[1:]); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "generate feature:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	flags := flag.NewFlagSet("feature", flag.ContinueOnError)
	root := flags.String("root", ".", "project root")
	if err := flags.Parse(args); err != nil {
		return err
	}

	if flags.NArg() != 1 {
		return errors.New("feature name is required: use feature <name>")
	}

	return generate(*root, flags.Arg(0))
}

func generate(root, name string) error {
	if !featureNamePattern.MatchString(name) {
		return errors.New("feature name must start with a lowercase letter and contain only lowercase letters, digits, or underscores")
	}

	dir := filepath.Join(root, "internal", "http", name)
	if _, err := os.Stat(dir); err == nil {
		return fmt.Errorf("feature %q already exists", name)
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("inspect feature directory: %w", err)
	}

	files := map[string]string{
		filepath.Join("dto", name+"_dto.go"): `// Package dto contains HTTP transport structures for the feature.
package dto

type ListResponse struct {
	Items []any ` + "`json:\"items\"`" + `
}
`,
		filepath.Join("handlers", name+"_handler.go"): `// Package handlers contains HTTP handlers for the feature.
package handlers

import (
	"fastgo/internal/http/` + name + `/services"
	"fastgo/internal/shared/response"

	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	service *services.Service
}

func New() *Handler {
	return &Handler{
		service: services.New(),
	}
}

func (h *Handler) List(c fiber.Ctx) error {
	return response.JSON(c, h.service.List())
}
`,
		filepath.Join("services", "service.go"): `package services

import "fastgo/internal/http/` + name + `/dto"

type Service struct{}

func New() *Service {
	return &Service{}
}

func (s *Service) List() dto.ListResponse {
	return dto.ListResponse{
		Items: []any{},
	}
}
`,
		"routes.go": fmt.Sprintf(`package %s

import (
	"github.com/gofiber/fiber/v3"

	"fastgo/internal/http/%s/handlers"
)

func RegisterRoutes(r fiber.Router) {
	%sGroup := r.Group("/%s")
	handler := handlers.New()

	%sGroup.Get("", handler.List)
}
`, name, name, name, name, name),
	}

	for path, content := range files {
		fullPath := filepath.Join(dir, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			return fmt.Errorf("create feature directory: %w", err)
		}

		if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
			return fmt.Errorf("write %s: %w", path, err)
		}
	}

	fmt.Printf("generated feature %s in %s\n", name, dir)
	return nil
}
