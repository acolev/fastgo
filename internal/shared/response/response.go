package response

import (
	"fastgo/internal/i18n"
	sharederrors "fastgo/internal/shared/errors"

	"github.com/gofiber/fiber/v3"
)

func JSON(c fiber.Ctx, data any) error {
	return c.JSON(fiber.Map{
		"data": data,
	})
}

type ErrorEnvelope struct {
	Error ErrorBody `json:"error"`
}

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

func ErrorHandler(c fiber.Ctx, err error) error {
	appErr := sharederrors.From(err)
	body := ErrorBody{
		Code:    appErr.Code,
		Message: i18n.T(c, appErr.MessageKey, appErr.Params),
	}

	if appErr.Details != nil {
		body.Details = appErr.Details
	}

	return c.Status(appErr.Status).JSON(ErrorEnvelope{
		Error: body,
	})
}

func Error(c fiber.Ctx, err error) error {
	return ErrorHandler(c, err)
}
