package handlers

import (
	"fastgo/internal/http/tests/dto"
	"fastgo/internal/http/tests/services"
	sharederrors "fastgo/internal/shared/errors"
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

func (h *Handler) CreateRange(c fiber.Ctx) error {
	var req dto.CreateRangeRequest
	if err := c.Bind().Body(&req); err != nil {
		return sharederrors.BadRequest("invalid_request_body", "errors.invalid_request_body", nil, nil).WithCause(err)
	}

	result, err := h.service.CreateRange(c.Context(), req.From, req.To)
	if err != nil {
		return err
	}

	return response.JSON(c, result)
}

func (h *Handler) List(c fiber.Ctx) error {
	result, err := h.service.List(c.Context())
	if err != nil {
		return err
	}

	return response.JSON(c, result)
}

func (h *Handler) Random(c fiber.Ctx) error {
	result, err := h.service.Random(c.Context())
	if err != nil {
		return err
	}

	return response.JSON(c, result)
}

func (h *Handler) Delete(c fiber.Ctx) error {
	result, err := h.service.Delete(c.Context(), c.Query("numbers"))
	if err != nil {
		return err
	}

	return response.JSON(c, result)
}

func (h *Handler) Clear(c fiber.Ctx) error {
	result, err := h.service.Clear(c.Context())
	if err != nil {
		return err
	}

	return response.JSON(c, result)
}
