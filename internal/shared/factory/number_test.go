package factory

import (
	"testing"

	"github.com/brianvoe/gofakeit/v7"
)

func TestNumberUsesRange(t *testing.T) {
	number := Number(gofakeit.New(42), NumberOptions{Min: 10, Max: 20})
	if number.Number < 10 || number.Number > 20 {
		t.Fatalf("expected number in range, got %d", number.Number)
	}
}

func TestNumberUsesOverride(t *testing.T) {
	override := 15
	number := Number(gofakeit.New(42), NumberOptions{
		Min:    10,
		Max:    20,
		Number: &override,
	})

	if number.Number != override {
		t.Fatalf("expected override %d, got %d", override, number.Number)
	}
}
