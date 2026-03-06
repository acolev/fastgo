package handlers

import (
	"fastgo/internal/http/probes/dto"

	"github.com/gofiber/fiber/v3"

	"fastgo/internal/http/probes/services"
)

type Handler struct {
	service *services.Service
}

var _ dto.ProbeResponse

func New() *Handler {
	return &Handler{
		service: services.New(),
	}
}

// Ping godoc
// @Summary Ping
// @Description Lightweight ping endpoint.
// @Tags probes
// @Produce json
// @Success 200 {object} dto.ProbeResponse
// @Router /ping [get]
func (h *Handler) Ping(c fiber.Ctx) error {
	return c.JSON(h.service.Ping())
}

// Health godoc
// @Summary Health
// @Description Basic liveness endpoint.
// @Tags probes
// @Produce json
// @Success 200 {object} dto.ProbeResponse
// @Router /health [get]
func (h *Handler) Health(c fiber.Ctx) error {
	return c.JSON(h.service.Health())
}

// Ready godoc
// @Summary Ready
// @Description Readiness endpoint that checks database and Redis connectivity.
// @Tags probes
// @Produce json
// @Success 200 {object} dto.ProbeResponse
// @Failure 503 {object} dto.ProbeResponse
// @Router /ready [get]
func (h *Handler) Ready(c fiber.Ctx) error {
	result := h.service.Ready()

	if result.Status != "ok" {
		return c.Status(fiber.StatusServiceUnavailable).JSON(result)
	}

	return c.JSON(result)
}
