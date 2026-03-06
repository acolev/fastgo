package dto

import "fastgo/internal/models"

type CreateRangeRequest struct {
	From int `json:"from"`
	To   int `json:"to"`
}

type CreateRangeResponse struct {
	From    int             `json:"from"`
	To      int             `json:"to"`
	Created int64           `json:"created"`
	Numbers []models.Number `json:"numbers"`
}

type ListResponse struct {
	Count   int             `json:"count"`
	Numbers []models.Number `json:"numbers"`
}

type DeleteResponse struct {
	Deleted int64 `json:"deleted"`
	Numbers []int `json:"numbers"`
}

type ClearResponse struct {
	Deleted int64 `json:"deleted"`
}
