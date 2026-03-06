package metrics

import (
	"github.com/gofiber/fiber/v3"

	"fastgo/internal/http/metrics/handlers"
)

func RegisterRoutes(r fiber.Router) {
	r.Get("/metrics", handlers.New())
}
