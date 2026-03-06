package services

import (
	"context"

	"fastgo/internal/http/probes/dto"
	"fastgo/internal/infra/database"
	appredis "fastgo/internal/infra/redis"
)

type Service struct{}

func New() *Service {
	return &Service{}
}

func (s *Service) Ping() dto.ProbeResponse {
	return dto.ProbeResponse{
		Status:  "ok",
		Message: "pong",
	}
}

func (s *Service) Health() dto.ProbeResponse {
	return dto.ProbeResponse{
		Status:  "ok",
		Message: "service is healthy",
	}
}

func (s *Service) Ready() dto.ProbeResponse {
	result := dto.ProbeResponse{
		Status:   "ok",
		Message:  "service is ready",
		Services: map[string]any{},
	}

	dbStatus := "ok"
	if err := database.DB().WithContext(context.Background()).Exec("SELECT 1").Error; err != nil {
		dbStatus = err.Error()
		result.Status = "error"
		result.Message = "service is not ready"
	}

	redisStatus := "ok"
	if err := appredis.Client().Ping(context.Background()).Err(); err != nil {
		redisStatus = err.Error()
		result.Status = "error"
		result.Message = "service is not ready"
	}

	result.Services["database"] = dbStatus
	result.Services["redis"] = redisStatus

	return result
}
