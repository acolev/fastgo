package dto

import "fastgo/internal/models"

type CreateRangeEnvelope struct {
	Data CreateRangeResponse `json:"data"`
}

type ListEnvelope struct {
	Data ListResponse `json:"data"`
}

type NumberEnvelope struct {
	Data models.Number `json:"data"`
}

type DeleteEnvelope struct {
	Data DeleteResponse `json:"data"`
}

type ClearEnvelope struct {
	Data ClearResponse `json:"data"`
}
