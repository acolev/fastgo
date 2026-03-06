package services

import (
	"context"
	crand "crypto/rand"
	"fmt"
	"math/big"
	"time"

	"fastgo/internal/http/tests/dto"
	"fastgo/internal/i18n"
	"fastgo/internal/infra/database"
	infraredis "fastgo/internal/infra/redis"
	"fastgo/internal/models"
	sharederrors "fastgo/internal/shared/errors"
	"fastgo/internal/shared/logger"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	minNumber           = 1
	maxNumber           = 199
	numbersListCacheTTL = 30 * time.Second
)

var numbersListCache = infraredis.NewJSONCache[dto.ListResponse]("tests:numbers", numbersListCacheTTL)

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
	}, s.invalidateCache(ctx)
}

func (s *Service) List(ctx context.Context) (dto.ListResponse, error) {
	return numbersListCache.Remember(ctx, "list", func(ctx context.Context) (dto.ListResponse, error) {
		var numbers []models.Number
		if err := database.DB().WithContext(ctx).Order("number ASC").Find(&numbers).Error; err != nil {
			return dto.ListResponse{}, fmt.Errorf("list numbers: %w", err)
		}

		return dto.ListResponse{
			Count:   len(numbers),
			Numbers: numbers,
		}, nil
	})
}

func (s *Service) Random(ctx context.Context) (models.Number, error) {
	list, err := s.List(ctx)
	if err != nil {
		return models.Number{}, err
	}

	if list.Count == 0 {
		return models.Number{}, sharederrors.NotFound("numbers_not_found", "errors.numbers_not_found", nil, nil)
	}

	offset, err := randomOffset(int64(list.Count))
	if err != nil {
		return models.Number{}, err
	}

	return list.Numbers[offset], nil
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
	}, s.invalidateCache(ctx)
}

func (s *Service) Clear(ctx context.Context) (dto.ClearResponse, error) {
	tx := database.DB().WithContext(ctx).
		Session(&gorm.Session{AllowGlobalUpdate: true}).
		Delete(&models.Number{})
	if tx.Error != nil {
		return dto.ClearResponse{}, fmt.Errorf("clear numbers: %w", tx.Error)
	}

	return dto.ClearResponse{Deleted: tx.RowsAffected}, s.invalidateCache(ctx)
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

func (s *Service) invalidateCache(ctx context.Context) error {
	if err := numbersListCache.InvalidateAll(ctx); err != nil {
		logger.Error(fmt.Sprintf("invalidate tests numbers cache: %v", err))
	}

	return nil
}
