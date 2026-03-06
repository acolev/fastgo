package bootstrap

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/gofiber/fiber/v3"
)

func loadSwaggerSpec(filePath, appName string) ([]byte, error) {
	raw, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read swagger spec: %w", err)
	}

	appName = strings.TrimSpace(appName)
	if appName == "" {
		return raw, nil
	}

	var spec map[string]any
	if err := json.Unmarshal(raw, &spec); err != nil {
		return nil, fmt.Errorf("decode swagger spec: %w", err)
	}

	info, ok := spec["info"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("swagger spec info section is invalid")
	}

	info["title"] = appName

	updated, err := json.Marshal(spec)
	if err != nil {
		return nil, fmt.Errorf("encode swagger spec: %w", err)
	}

	return updated, nil
}

func swaggerSpecHandler(spec []byte) fiber.Handler {
	return func(c fiber.Ctx) error {
		c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSONCharsetUTF8)
		c.Set(fiber.HeaderCacheControl, "no-store")

		return c.Send(spec)
	}
}
