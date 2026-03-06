package errors

import (
	stderrs "errors"

	"fastgo/internal/i18n"

	"github.com/gofiber/fiber/v3"
)

var (
	ErrNotFound     = stderrs.New("not found")
	ErrUnauthorized = stderrs.New("unauthorized")
)

type AppError struct {
	Status     int
	Code       string
	MessageKey string
	Params     i18n.Params
	Details    any
	Err        error
}

func New(status int, code, messageKey string, params i18n.Params, details any) *AppError {
	return &AppError{
		Status:     status,
		Code:       code,
		MessageKey: messageKey,
		Params:     params,
		Details:    details,
	}
}

func BadRequest(code, messageKey string, params i18n.Params, details any) *AppError {
	return New(fiber.StatusBadRequest, code, messageKey, params, details)
}

func Unauthorized(code, messageKey string, params i18n.Params, details any) *AppError {
	return New(fiber.StatusUnauthorized, code, messageKey, params, details)
}

func Forbidden(code, messageKey string, params i18n.Params, details any) *AppError {
	return New(fiber.StatusForbidden, code, messageKey, params, details)
}

func NotFound(code, messageKey string, params i18n.Params, details any) *AppError {
	return New(fiber.StatusNotFound, code, messageKey, params, details)
}

func Internal(code, messageKey string, params i18n.Params, details any) *AppError {
	return New(fiber.StatusInternalServerError, code, messageKey, params, details)
}

func (e *AppError) Error() string {
	if e == nil {
		return ""
	}

	if e.Err != nil {
		return e.Err.Error()
	}

	if e.Code != "" {
		return e.Code
	}

	return e.MessageKey
}

func (e *AppError) Unwrap() error {
	if e == nil {
		return nil
	}

	return e.Err
}

func (e *AppError) WithCause(err error) *AppError {
	e.Err = err
	return e
}

func From(err error) *AppError {
	if err == nil {
		return Internal("internal_error", "errors.internal", nil, nil)
	}

	var appErr *AppError
	if stderrs.As(err, &appErr) {
		return appErr
	}

	var fiberErr *fiber.Error
	if stderrs.As(err, &fiberErr) {
		return fromFiberError(fiberErr)
	}

	return Internal("internal_error", "errors.internal", nil, nil).WithCause(err)
}

func fromFiberError(err *fiber.Error) *AppError {
	switch err.Code {
	case fiber.StatusBadRequest:
		return BadRequest("bad_request", "errors.bad_request", nil, nil).WithCause(err)
	case fiber.StatusUnauthorized:
		return Unauthorized("unauthorized", "errors.unauthorized", nil, nil).WithCause(err)
	case fiber.StatusForbidden:
		return Forbidden("forbidden", "errors.forbidden", nil, nil).WithCause(err)
	case fiber.StatusNotFound:
		return NotFound("not_found", "errors.not_found", nil, nil).WithCause(err)
	case fiber.StatusMethodNotAllowed:
		return New(fiber.StatusMethodNotAllowed, "method_not_allowed", "errors.method_not_allowed", nil, nil).WithCause(err)
	default:
		if err.Code >= fiber.StatusInternalServerError {
			return Internal("internal_error", "errors.internal", nil, nil).WithCause(err)
		}

		return New(err.Code, "request_error", "errors.bad_request", nil, nil).WithCause(err)
	}
}
