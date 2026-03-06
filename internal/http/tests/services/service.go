package services

import (
	"context"
	crand "crypto/rand"
	"errors"
	"fmt"
	"math/big"

	"fastgo/internal/http/tests/dto"
	"fastgo/internal/i18n"
	"fastgo/internal/infra/database"
	"fastgo/internal/models"
	sharederrors "fastgo/internal/shared/errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	minNumber = 1
	maxNumber = 199
)

type Service struct{}

func New() *Service {
	return &Service{}
}

func (s *Service) CreateRange(ctx context.Context, from, to int) (dto.CreateRangeResponse, error) {
	if err := validateRange(from, to); err != nil {
		return dto.CreateRangeResponse{}, err
	}

	numbers := make([]models.Number, 0, to-from+1)
	for value := from; value <= to; value++ {
		numbers = append(numbers, models.Number{Number: value})
	}

	tx := database.DB().WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "number"}},
			DoNothing: true,
		}).
		Create(&numbers)
	if tx.Error != nil {
		return dto.CreateRangeResponse{}, fmt.Errorf("create numbers range: %w", tx.Error)
	}

	var stored []models.Number
	if err := database.DB().WithContext(ctx).
		Where("number BETWEEN ? AND ?", from, to).
		Order("number ASC").
		Find(&stored).Error; err != nil {
		return dto.CreateRangeResponse{}, fmt.Errorf("list created numbers range: %w", err)
	}

	return dto.CreateRangeResponse{
		From:    from,
		To:      to,
		Created: tx.RowsAffected,
		Numbers: stored,
	}, nil
}

func (s *Service) List(ctx context.Context) (dto.ListResponse, error) {
	var numbers []models.Number
	if err := database.DB().WithContext(ctx).Order("number ASC").Find(&numbers).Error; err != nil {
		return dto.ListResponse{}, fmt.Errorf("list numbers: %w", err)
	}

	return dto.ListResponse{
		Count:   len(numbers),
		Numbers: numbers,
	}, nil
}

func (s *Service) Random(ctx context.Context) (models.Number, error) {
	var count int64
	if err := database.DB().WithContext(ctx).Model(&models.Number{}).Count(&count).Error; err != nil {
		return models.Number{}, fmt.Errorf("count numbers: %w", err)
	}

	if count == 0 {
		return models.Number{}, sharederrors.NotFound("numbers_not_found", "errors.numbers_not_found", nil, nil)
	}

	offset, err := randomOffset(count)
	if err != nil {
		return models.Number{}, err
	}

	var number models.Number
	if err := database.DB().WithContext(ctx).
		Order("number ASC").
		Offset(offset).
		Limit(1).
		First(&number).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.Number{}, sharederrors.NotFound("numbers_not_found", "errors.numbers_not_found", nil, nil)
		}

		return models.Number{}, fmt.Errorf("random number: %w", err)
	}

	return number, nil
}

func (s *Service) Delete(ctx context.Context, rawNumbers string) (dto.DeleteResponse, error) {
	numbers, err := parseNumbersCSV(rawNumbers)
	if err != nil {
		return dto.DeleteResponse{}, err
	}

	tx := database.DB().WithContext(ctx).Where("number IN ?", numbers).Delete(&models.Number{})
	if tx.Error != nil {
		return dto.DeleteResponse{}, fmt.Errorf("delete numbers: %w", tx.Error)
	}

	return dto.DeleteResponse{
		Deleted: tx.RowsAffected,
		Numbers: numbers,
	}, nil
}

func (s *Service) Clear(ctx context.Context) (dto.ClearResponse, error) {
	tx := database.DB().WithContext(ctx).
		Session(&gorm.Session{AllowGlobalUpdate: true}).
		Delete(&models.Number{})
	if tx.Error != nil {
		return dto.ClearResponse{}, fmt.Errorf("clear numbers: %w", tx.Error)
	}

	return dto.ClearResponse{Deleted: tx.RowsAffected}, nil
}

func validateRange(from, to int) error {
	if from < minNumber || from > maxNumber || to < minNumber || to > maxNumber {
		return sharederrors.BadRequest(
			"numbers_range_between",
			"errors.numbers_range_between",
			i18n.Params{
				"min": minNumber,
				"max": maxNumber,
			},
			map[string]any{
				"min": minNumber,
				"max": maxNumber,
			},
		)
	}

	if from > to {
		return sharederrors.BadRequest(
			"numbers_range_order",
			"errors.numbers_range_order",
			nil,
			map[string]any{
				"from": from,
				"to":   to,
			},
		)
	}

	return nil
}

func randomOffset(max int64) (int, error) {
	if max <= 0 {
		return 0, sharederrors.NotFound("numbers_not_found", "errors.numbers_not_found", nil, nil)
	}

	value, err := crand.Int(crand.Reader, big.NewInt(max))
	if err != nil {
		return 0, fmt.Errorf("generate random offset: %w", err)
	}

	return int(value.Int64()), nil
}
