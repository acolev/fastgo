package probes

import (
	"github.com/gofiber/fiber/v3"

	"fastgo/internal/http/probes/handlers"
)

func RegisterRoutes(r fiber.Router) {
	handler := handlers.New()

	r.Get("/ping", handler.Ping)
	r.Get("/health", handler.Health)
	r.Get("/ready", handler.Ready)
}
