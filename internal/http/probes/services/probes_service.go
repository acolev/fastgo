package services

import (
	"context"
	"time"

	"fastgo/internal/http/probes/dto"
	"fastgo/internal/infra/database"
	appredis "fastgo/internal/infra/redis"
)

const readinessTimeout = 3 * time.Second

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
	dbCtx, dbCancel := context.WithTimeout(context.Background(), readinessTimeout)
	if err := database.Ping(dbCtx); err != nil {
		dbStatus = err.Error()
		result.Status = "error"
		result.Message = "service is not ready"
	}
	dbCancel()

	redisStatus := "ok"
	redisCtx, redisCancel := context.WithTimeout(context.Background(), readinessTimeout)
	if err := appredis.Ping(redisCtx); err != nil {
		redisStatus = err.Error()
		result.Status = "error"
		result.Message = "service is not ready"
	}
	redisCancel()

	result.Services["database"] = dbStatus
	result.Services["redis"] = redisStatus

	return result
}
