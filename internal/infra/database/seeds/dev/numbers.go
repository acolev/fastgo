package dev

import (
	"context"
	"fmt"

	"fastgo/internal/models"
	"fastgo/internal/shared/factory"

	"github.com/brianvoe/gofakeit/v7"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	NumbersName     = "dev.numbers"
	minSeededNumber = 10_000
	maxSeededNumber = 99_999
)

type Numbers struct {
	count int
	faker *gofakeit.Faker
}

func NewNumbers(seed uint64, count int) (*Numbers, error) {
	if count < 1 || count > maxSeededNumber-minSeededNumber+1 {
		return nil, fmt.Errorf("count must be between 1 and %d", maxSeededNumber-minSeededNumber+1)
	}

	return &Numbers{
		count: count,
		faker: gofakeit.New(seed),
	}, nil
}

func (s *Numbers) Name() string {
	return NumbersName
}

func (s *Numbers) Run(ctx context.Context, tx *gorm.DB) error {
	numbers := s.generate()

	if err := tx.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "number"}},
			DoNothing: true,
		}).
		Create(&numbers).Error; err != nil {
		return fmt.Errorf("create numbers: %w", err)
	}

	return nil
}

func (s *Numbers) Reset(ctx context.Context, tx *gorm.DB) error {
	if err := tx.WithContext(ctx).
		Where("number BETWEEN ? AND ?", minSeededNumber, maxSeededNumber).
		Delete(&models.Number{}).Error; err != nil {
		return fmt.Errorf("delete numbers: %w", err)
	}

	return nil
}

func (s *Numbers) generate() []models.Number {
	numbers := make([]models.Number, 0, s.count)
	seen := make(map[int]struct{}, s.count)

	for len(numbers) < s.count {
		number := factory.Number(s.faker, factory.NumberOptions{
			Min: minSeededNumber,
			Max: maxSeededNumber,
		})
		if _, exists := seen[number.Number]; exists {
			continue
		}

		seen[number.Number] = struct{}{}
		numbers = append(numbers, number)
	}

	return numbers
}
