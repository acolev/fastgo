package services

import (
	"sort"
	"strconv"
	"strings"

	"fastgo/internal/i18n"
	sharederrors "fastgo/internal/shared/errors"
)

func parseNumbersCSV(raw string) ([]int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, sharederrors.BadRequest(
			"numbers_query_required",
			"errors.numbers_query_required",
			nil,
			map[string]any{
				"parameter": "numbers",
			},
		)
	}

	parts := strings.Split(raw, ",")
	numbers := make([]int, 0, len(parts))
	seen := make(map[int]struct{}, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			return nil, sharederrors.BadRequest(
				"numbers_query_empty_values",
				"errors.numbers_query_empty_values",
				nil,
				map[string]any{
					"parameter": "numbers",
				},
			)
		}

		number, err := strconv.Atoi(part)
		if err != nil {
			return nil, sharederrors.BadRequest(
				"numbers_invalid_value",
				"errors.numbers_invalid_value",
				i18n.Params{
					"value": part,
				},
				map[string]any{
					"value": part,
				},
			)
		}

		if number < minNumber || number > maxNumber {
			return nil, sharederrors.BadRequest(
				"numbers_out_of_range",
				"errors.numbers_out_of_range",
				i18n.Params{
					"value": number,
					"min":   minNumber,
					"max":   maxNumber,
				},
				map[string]any{
					"value": number,
					"min":   minNumber,
					"max":   maxNumber,
				},
			)
		}

		if _, exists := seen[number]; exists {
			continue
		}

		seen[number] = struct{}{}
		numbers = append(numbers, number)
	}

	sort.Ints(numbers)

	return numbers, nil
}
