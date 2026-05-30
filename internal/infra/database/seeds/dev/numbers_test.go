package dev

import (
	"reflect"
	"testing"
)

func TestNumbersGenerateIsReproducibleAndUnique(t *testing.T) {
	first, err := NewNumbers(42, 25)
	if err != nil {
		t.Fatalf("NewNumbers returned error: %v", err)
	}

	second, err := NewNumbers(42, 25)
	if err != nil {
		t.Fatalf("NewNumbers returned error: %v", err)
	}

	firstNumbers := first.generate()
	secondNumbers := second.generate()
	if !reflect.DeepEqual(firstNumbers, secondNumbers) {
		t.Fatal("expected the same faker seed to generate the same numbers")
	}

	seen := make(map[int]struct{}, len(firstNumbers))
	for _, number := range firstNumbers {
		if number.Number < minSeededNumber || number.Number > maxSeededNumber {
			t.Fatalf("generated number %d is outside the dev seed range", number.Number)
		}

		if _, exists := seen[number.Number]; exists {
			t.Fatalf("generated duplicate number %d", number.Number)
		}

		seen[number.Number] = struct{}{}
	}
}

func TestNewNumbersRejectsInvalidCount(t *testing.T) {
	if _, err := NewNumbers(42, 0); err == nil {
		t.Fatal("expected invalid count error")
	}
}
