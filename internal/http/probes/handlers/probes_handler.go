package handlers

import (
	"github.com/gofiber/fiber/v3"

	"fastgo/internal/http/probes/services"
)

type Handler struct {
	service *services.Service
}

func New() *Handler {
	return &Handler{
		service: services.New(),
	}
}

func (h *Handler) Ping(c fiber.Ctx) error {
	return c.JSON(h.service.Ping())
}

func (h *Handler) Health(c fiber.Ctx) error {
	return c.JSON(h.service.Health())
}

func (h *Handler) Ready(c fiber.Ctx) error {
	result := h.service.Ready()

	if result.Status != "ok" {
		return c.Status(fiber.StatusServiceUnavailable).JSON(result)
	}

	return c.JSON(result)
}
