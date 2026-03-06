package services

import (
	"errors"
	"testing"

	sharederrors "fastgo/internal/shared/errors"

	"github.com/gofiber/fiber/v3"
)

func TestCreateRangeRejectsOutOfBounds(t *testing.T) {
	err := validateRange(0, 10)
	if err == nil {
		t.Fatal("expected validation error")
	}

	var appErr *sharederrors.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppError, got %T", err)
	}

	if appErr.Status != fiber.StatusBadRequest {
		t.Fatalf("got status %d, want %d", appErr.Status, fiber.StatusBadRequest)
	}

	if appErr.Code != "numbers_range_between" {
		t.Fatalf("got code %q, want %q", appErr.Code, "numbers_range_between")
	}
}

func TestCreateRangeRejectsDescendingRange(t *testing.T) {
	err := validateRange(10, 5)
	if err == nil {
		t.Fatal("expected validation error")
	}

	var appErr *sharederrors.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppError, got %T", err)
	}

	if appErr.Code != "numbers_range_order" {
		t.Fatalf("got code %q, want %q", appErr.Code, "numbers_range_order")
	}
}

func TestDeleteParsesNumbersCSV(t *testing.T) {
	result, err := parseNumbersCSV("5, 1, 5, 8")
	if err != nil {
		t.Fatalf("parseNumbersCSV returned error: %v", err)
	}

	want := []int{1, 5, 8}
	if len(result) != len(want) {
		t.Fatalf("got %v, want %v", result, want)
	}

	for i := range want {
		if result[i] != want[i] {
			t.Fatalf("got %v, want %v", result, want)
		}
	}
}

func TestDeleteRejectsInvalidCSV(t *testing.T) {
	_, err := parseNumbersCSV("1,abc")
	if err == nil {
		t.Fatal("expected validation error")
	}

	var appErr *sharederrors.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppError, got %T", err)
	}

	if appErr.Code != "numbers_invalid_value" {
		t.Fatalf("got code %q, want %q", appErr.Code, "numbers_invalid_value")
	}
}
