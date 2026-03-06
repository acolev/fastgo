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

// CreateRange godoc
// @Summary Create a range of numbers
// @Description Creates numbers in the inclusive range from `from` to `to`. Allowed range is 1..199.
// @Tags tests/numbers
// @Accept json
// @Produce json
// @Param request body dto.CreateRangeRequest true "Range payload"
// @Success 200 {object} dto.CreateRangeEnvelope
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /v1/t/numbers/range [post]
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

// List godoc
// @Summary List numbers
// @Description Returns all stored numbers ordered ascending.
// @Tags tests/numbers
// @Produce json
// @Success 200 {object} dto.ListEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /v1/t/numbers [get]
func (h *Handler) List(c fiber.Ctx) error {
	result, err := h.service.List(c.Context())
	if err != nil {
		return err
	}

	return response.JSON(c, result)
}

// Random godoc
// @Summary Get random number
// @Description Returns one random stored number.
// @Tags tests/numbers
// @Produce json
// @Success 200 {object} dto.NumberEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /v1/t/numbers/random [get]
func (h *Handler) Random(c fiber.Ctx) error {
	result, err := h.service.Random(c.Context())
	if err != nil {
		return err
	}

	return response.JSON(c, result)
}

// Delete godoc
// @Summary Delete numbers
// @Description Deletes numbers passed in the `numbers` query parameter as comma-separated values.
// @Tags tests/numbers
// @Produce json
// @Param numbers query string true "Comma-separated numbers" example(1,2,3)
// @Success 200 {object} dto.DeleteEnvelope
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /v1/t/numbers [delete]
func (h *Handler) Delete(c fiber.Ctx) error {
	result, err := h.service.Delete(c.Context(), c.Query("numbers"))
	if err != nil {
		return err
	}

	return response.JSON(c, result)
}

// Clear godoc
// @Summary Clear numbers
// @Description Deletes all stored numbers.
// @Tags tests/numbers
// @Produce json
// @Success 200 {object} dto.ClearEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /v1/t/numbers/clear [delete]
func (h *Handler) Clear(c fiber.Ctx) error {
	result, err := h.service.Clear(c.Context())
	if err != nil {
		return err
	}

	return response.JSON(c, result)
}
