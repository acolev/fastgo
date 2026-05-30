package factory

import (
	"fastgo/internal/models"

	"github.com/brianvoe/gofakeit/v7"
)

type NumberOptions struct {
	Min    int
	Max    int
	Number *int
}

func Number(faker *gofakeit.Faker, opts NumberOptions) models.Number {
	number := faker.Number(opts.Min, opts.Max)
	if opts.Number != nil {
		number = *opts.Number
	}

	return models.Number{Number: number}
}
