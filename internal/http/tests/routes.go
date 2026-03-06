package tests

import (
	"github.com/gofiber/fiber/v3"

	"fastgo/internal/http/tests/handlers"
)

func RegisterRoutes(r fiber.Router) {
	numbersGroup := r.Group("/numbers")
	handler := handlers.New()

	numbersGroup.Post("/range", handler.CreateRange)
	numbersGroup.Get("", handler.List)
	numbersGroup.Get("/random", handler.Random)
	numbersGroup.Delete("", handler.Delete)
	numbersGroup.Delete("/clear", handler.Clear)
}
